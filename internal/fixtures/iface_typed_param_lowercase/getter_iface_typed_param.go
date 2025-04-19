package iface_typed_param_lowercase

type GetterIfaceTypedParam[a comparable] interface {
	Get(a) a
}
