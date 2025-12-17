package nodes

type StringifyArg = map[string]interface{}
type StringifyNode interface {
	Stringify(StringifyArg) string
}
