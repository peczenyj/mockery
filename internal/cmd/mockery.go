package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/vektra/mockery/v3/config"
	"github.com/vektra/mockery/v3/internal"
	pkg "github.com/vektra/mockery/v3/internal"
	"github.com/vektra/mockery/v3/internal/logging"
	"github.com/vektra/mockery/v3/internal/stackerr"

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/tools/go/packages"
)

var ErrCfgFileNotFound = errors.New("config file not found")

func NewRootCmd() (*cobra.Command, error) {
	var pFlags *pflag.FlagSet
	cmd := &cobra.Command{
		Use:   "mockery",
		Short: "Generate mock objects for your Go interfaces",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pFlags.Parse(args); err != nil {
				fmt.Printf("failed to parse flags: %s", err.Error())
				os.Exit(1)
			}
			log, err := logging.GetLogger("info")
			if err != nil {
				fmt.Printf("failed to get logger: %s\n", err.Error())
				os.Exit(1)
			}
			ctx := log.WithContext(context.Background())

			r, err := GetRootApp(ctx, pFlags)
			if err != nil {
				logFatalErr(ctx, err)
			}
			if err := r.Run(); err != nil {
				logFatalErr(ctx, err)
			}
		},
	}
	pFlags = cmd.PersistentFlags()
	pFlags.String("config", "", "config file to use")
	pFlags.String("log-level", os.Getenv("MOCKERY_LOG_LEVEL"), "Level of logging")

	cmd.AddCommand(NewShowConfigCmd())
	cmd.AddCommand(NewVersionCmd())
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewMigrateCmd())
	return cmd, nil
}

func logFatalErr(ctx context.Context, err error) {
	log := zerolog.Ctx(ctx)
	log.Fatal().Err(err).Msg("app failed")
}

// Execute executes the cobra CLI workflow
func Execute() {
	cmd, err := NewRootCmd()
	if err != nil {
		os.Exit(1)
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type RootApp struct {
	Config config.RootConfig
}

func GetRootApp(ctx context.Context, flags *pflag.FlagSet) (*RootApp, error) {
	r := &RootApp{}
	config, _, err := config.NewRootConfig(ctx, flags)
	if err != nil {
		return nil, fmt.Errorf("getting config: %w", err)
	}
	r.Config = *config
	return r, nil
}

// InterfaceCollection maintains a list of *pkg.Interface and asserts that all
// the interfaces in the collection belong to the same source package. It also
// asserts that various properties of the interfaces added to the collection are
// uniform.
type InterfaceCollection struct {
	// Mockery needs to assert that certain properties of the added interfaces
	// are uniform for all members of the collection. This includes things like
	// 1. Package name of the output mock file
	// 2. Source package path (only one package per output file is allowed)
	srcPkgPath  string
	outFilePath *pathlib.Path
	srcPkg      *packages.Package
	outPkgName  string
	interfaces  []*internal.Interface
	template    string
}

func NewInterfaceCollection(
	srcPkgPath string,
	outFilePath *pathlib.Path,
	srcPkg *packages.Package,
	outPkgName string,
	templ string,
) *InterfaceCollection {
	return &InterfaceCollection{
		srcPkgPath:  srcPkgPath,
		outFilePath: outFilePath,
		srcPkg:      srcPkg,
		outPkgName:  outPkgName,
		interfaces:  make([]*internal.Interface, 0),
		template:    templ,
	}
}

func (i *InterfaceCollection) Append(ctx context.Context, iface *internal.Interface) error {
	collectionFilepath := i.outFilePath.String()
	interfaceFilepath := iface.Config.FilePath().String()
	log := zerolog.Ctx(ctx).With().
		Str(logging.LogKeyInterface, iface.Name).
		Str("collection-pkgname", i.outPkgName).
		Str("interface-pkgname", *iface.Config.PkgName).
		Str("collection-pkgpath", i.srcPkgPath).
		Str("interface-pkgpath", iface.Pkg.PkgPath).
		Str("collection-filepath", collectionFilepath).
		Str("interface-filepath", interfaceFilepath).
		Logger()

	if collectionFilepath != interfaceFilepath {
		msg := "all mocks in an InterfaceCollection must have the same output file path"
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	if i.outPkgName != *iface.Config.PkgName {
		msg := "all mocks in an output file must have the same pkgname"
		log.Error().Str("interface-pkgname", *iface.Config.PkgName).Msg(msg)
		return errors.New(msg)
	}
	if i.srcPkgPath != iface.Pkg.PkgPath {
		msg := "all mocks in an output file must come from the same source package"
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	if i.template != *iface.Config.Template {
		msg := "all mocks in an output file must use the same template"
		log.Error().Str("expected-template", i.template).Str("interface-template", *iface.Config.Template).Msg(msg)
		return errors.New(msg)
	}
	i.interfaces = append(i.interfaces, iface)
	return nil
}

func (r *RootApp) Run() error {
	remoteTemplateCache := make(map[string]*internal.RemoteTemplate)

	log, err := logging.GetLogger(*r.Config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}
	log.Info().Str("config-file", r.Config.ConfigFileUsed().String()).Msgf("Starting mockery")
	ctx := log.WithContext(context.Background())

	if err := r.Config.Initialize(ctx); err != nil {
		return err
	}

	buildTags := strings.Split(*r.Config.BuildTags, " ")

	configuredPackages, err := r.Config.GetPackages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get package from config: %w", err)
	}
	if len(configuredPackages) == 0 {
		log.Error().Msg("no packages specified in config")
		return fmt.Errorf("no packages specified in config")
	}
	parser := pkg.NewParser(buildTags)

	// Let's build a missing map here to keep track of seen interfaces.
	// (pkg -> list of interface names)
	// After seeing an interface it'll be deleted from the map, keeping only
	// missing interfaces or packages in there.
	//
	// NOTE: We do that here without relying on parser, because parses iterates
	// over existing go files and interfaces, while user could've had a typo in
	// interface or pacakge name, making it impossible for parser to find these
	// files/interfaces in the first place.
	log.Debug().Msg("Making seen map...")
	missingMap := make(map[string]map[string]struct{}, len(configuredPackages))
	for _, p := range configuredPackages {
		config, err := r.Config.GetPackageConfig(ctx, p)
		if err != nil {
			return err
		}
		if _, ok := missingMap[p]; !ok {
			missingMap[p] = make(map[string]struct{}, len(config.Interfaces))
		}

		for ifaceName := range config.Interfaces {
			missingMap[p][ifaceName] = struct{}{}
		}
	}

	log.Info().Msg("Parsing configured packages...")
	interfaces, err := parser.ParsePackages(ctx, configuredPackages)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse packages")
		return err
	}
	log.Info().Msg("Done parsing configured packages.")
	// maps the following:
	// outputFilePath|fullyQualifiedInterfaceName|[]*pkg.Interface
	// The reason why we need an interior map of fully qualified interface name
	// to a slice of *pkg.Interface (which represents all information necessary
	// to create the output mock) is because mockery allows multiple mocks to
	// be created for each input interface.
	mockFileToInterfaces := map[string]*InterfaceCollection{}

	for _, iface := range interfaces {
		ifaceLog := log.
			With().
			Str(logging.LogKeyInterface, iface.Name).
			Str(logging.LogKeyPackagePath, iface.Pkg.Types.Path()).
			Logger()

		if _, exist := missingMap[iface.Pkg.PkgPath]; exist {
			delete(missingMap[iface.Pkg.PkgPath], iface.Name)

			if len(missingMap[iface.Pkg.PkgPath]) == 0 {
				delete(missingMap, iface.Pkg.PkgPath)
			}
		}

		ifaceCtx := ifaceLog.WithContext(ctx)

		pkgConfig, err := r.Config.GetPackageConfig(ctx, iface.Pkg.PkgPath)
		if err != nil {
			return fmt.Errorf("getting package %s: %w", iface.Pkg.PkgPath, err)
		}
		ifaceLog.Debug().Str("root-mock-name", *r.Config.Config.StructName).Str("pkg-mock-name", *pkgConfig.Config.StructName).Msg("mock-name during first GetPackageConfig")

		shouldGenerate, err := pkgConfig.ShouldGenerateInterface(ifaceCtx, iface.Name)
		if err != nil {
			return err
		}
		if !shouldGenerate {
			ifaceLog.Debug().Msg("config doesn't specify to generate this interface, skipping")
			continue
		}
		if pkgConfig.Interfaces == nil {
			ifaceLog.Debug().Msg("interfaces is nil")
		}
		ifaceConfig := pkgConfig.GetInterfaceConfig(ctx, iface.Name)
		for _, ifaceConfig := range ifaceConfig.Configs {
			if err := ifaceConfig.ParseTemplates(ifaceCtx, iface.FileName, iface.Name, iface.Pkg); err != nil {
				log.Err(err).Msg("Can't parse config templates for interface")
				return err
			}
			filePath := ifaceConfig.FilePath().Clean()
			ifaceLog.Info().Str("collection", filePath.String()).Msg("adding interface to collection")

			_, ok := mockFileToInterfaces[filePath.String()]
			if !ok {
				mockFileToInterfaces[filePath.String()] = NewInterfaceCollection(
					iface.Pkg.PkgPath,
					filePath,
					iface.Pkg,
					*ifaceConfig.PkgName,
					*ifaceConfig.Template,
				)
			}
			if err := mockFileToInterfaces[filePath.String()].Append(
				ctx,
				internal.NewInterface(
					iface.Name,
					iface.TypeSpec,
					iface.GenDecl,
					iface.FileName,
					iface.File,
					iface.Pkg,
					ifaceConfig),
			); err != nil {
				return err
			}
		}
	}

	for outFilePath, interfacesInFile := range mockFileToInterfaces {
		fileLog := log.With().Str("file", outFilePath).Logger()
		fileCtx := fileLog.WithContext(ctx)

		fileLog.Debug().Int("interfaces-in-file-len", len(interfacesInFile.interfaces)).Msgf("%v", interfacesInFile)

		packageConfig, err := r.Config.GetPackageConfig(fileCtx, interfacesInFile.srcPkgPath)
		if err != nil {
			return err
		}
		if err := packageConfig.Config.ParseTemplates(ctx, "", "", interfacesInFile.srcPkg); err != nil {
			return err
		}

		generator, err := pkg.NewTemplateGenerator(
			fileCtx,
			interfacesInFile.srcPkg,
			interfacesInFile.outFilePath.Parent(),
			*packageConfig.Config.Template,
			*packageConfig.Config.TemplateSchema,
			*packageConfig.Config.RequireTemplateSchemaExists,
			remoteTemplateCache,
			pkg.Formatter(*r.Config.Formatter),
			packageConfig.Config,
			interfacesInFile.outPkgName,
		)
		if err != nil {
			return err
		}
		fileLog.Info().Msg("Executing template")
		templateBytes, err := generator.Generate(fileCtx, interfacesInFile.interfaces)
		if err != nil {
			return err
		}

		outFile := pathlib.NewPath(outFilePath)
		if err := outFile.Parent().MkdirAll(); err != nil {
			log.Err(err).Msg("failed to mkdir parent directories of mock file")
			return stackerr.NewStackErr(err)
		}
		fileLog.Info().Msg("Writing template to file")
		outFileExists, err := outFile.Exists()
		if err != nil {
			fileLog.Err(err).Msg("can't determine if outfile exists")
			return fmt.Errorf("determining if outfile exists: %w", err)
		}
		if outFileExists && !*packageConfig.Config.ForceFileWrite {
			fileLog.Error().Bool("force-file-write", *packageConfig.Config.ForceFileWrite).Msg("output file exists, can't write mocks")
			return fmt.Errorf("outfile exists")
		}

		if err := outFile.WriteFile(templateBytes); err != nil {
			return stackerr.NewStackErr(err)
		}
	}

	// The loop above could exit early, so sometimes warnings won't be shown
	// until other errors are fixed
	var foundMissing bool
	for packagePath := range missingMap {
		for ifaceName := range missingMap[packagePath] {
			foundMissing = true
			log.Error().
				Str(logging.LogKeyInterface, ifaceName).
				Str(logging.LogKeyPackagePath, packagePath).
				Msg("interface not found in source")
		}
	}
	if foundMissing {
		os.Exit(1)
	}

	return nil
}
