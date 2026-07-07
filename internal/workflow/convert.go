package workflow

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// toStarlark converts a Go value to a Starlark value. It supports the value
// shapes produced by JSON and YAML decoding.
func toStarlark(v interface{}) (starlark.Value, error) {
	switch val := v.(type) {
	case nil:
		return starlark.None, nil
	case bool:
		return starlark.Bool(val), nil
	case int:
		return starlark.MakeInt(val), nil
	case int64:
		return starlark.MakeInt64(val), nil
	case float64:
		return starlark.Float(val), nil
	case string:
		return starlark.String(val), nil
	case []interface{}:
		items := make([]starlark.Value, len(val))
		for i, item := range val {
			sv, err := toStarlark(item)
			if err != nil {
				return nil, err
			}
			items[i] = sv
		}
		return starlark.NewList(items), nil
	case []string:
		items := make([]starlark.Value, len(val))
		for i, item := range val {
			items[i] = starlark.String(item)
		}
		return starlark.NewList(items), nil
	case map[string]interface{}:
		dict := starlark.NewDict(len(val))
		for k, item := range val {
			sv, err := toStarlark(item)
			if err != nil {
				return nil, err
			}
			if err := dict.SetKey(starlark.String(k), sv); err != nil {
				return nil, err
			}
		}
		return dict, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", v)
	}
}

// fromStarlark converts a Starlark value to a Go value suitable for JSON
// encoding.
func fromStarlark(v starlark.Value) (interface{}, error) {
	switch val := v.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(val), nil
	case starlark.Int:
		i, ok := val.Int64()
		if !ok {
			return nil, fmt.Errorf("integer result too large: %s", val)
		}
		return i, nil
	case starlark.Float:
		return float64(val), nil
	case starlark.String:
		return string(val), nil
	case *starlark.List:
		return fromStarlarkSequence(val.Len(), val.Index)
	case starlark.Tuple:
		return fromStarlarkSequence(val.Len(), val.Index)
	case *starlark.Set:
		items := make([]interface{}, 0, val.Len())
		iter := val.Iterate()
		defer iter.Done()
		var elem starlark.Value
		for iter.Next(&elem) {
			gv, err := fromStarlark(elem)
			if err != nil {
				return nil, err
			}
			items = append(items, gv)
		}
		return items, nil
	case *starlark.Dict:
		out := make(map[string]interface{}, val.Len())
		for _, item := range val.Items() {
			key, ok := item[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("dict result keys must be strings, got %s", item[0].Type())
			}
			gv, err := fromStarlark(item[1])
			if err != nil {
				return nil, err
			}
			out[string(key)] = gv
		}
		return out, nil
	case *starlarkstruct.Struct:
		out := make(map[string]interface{})
		for _, name := range val.AttrNames() {
			attr, err := val.Attr(name)
			if err != nil {
				return nil, err
			}
			gv, err := fromStarlark(attr)
			if err != nil {
				return nil, err
			}
			out[name] = gv
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported result type %s", v.Type())
	}
}

func fromStarlarkSequence(n int, index func(int) starlark.Value) (interface{}, error) {
	items := make([]interface{}, n)
	for i := 0; i < n; i++ {
		gv, err := fromStarlark(index(i))
		if err != nil {
			return nil, err
		}
		items[i] = gv
	}
	return items, nil
}
