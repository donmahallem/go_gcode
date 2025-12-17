package nodes

import (
	"fmt"
)

func evalPrefixExpression(operator string, right any) (any, error) {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return nil, fmt.Errorf("unknown operator: %s%T", operator, right)
	}
}

func evalBangOperatorExpression(right any) (any, error) {
	switch right {
	case true:
		return false, nil
	case false:
		return true, nil
	case nil:
		return true, nil
	default:
		return false, nil
	}
}

func evalMinusPrefixOperatorExpression(right any) (any, error) {
	switch val := right.(type) {
	case int:
		return -int64(val), nil
	case int64:
		return -val, nil
	case float64:
		return -val, nil
	default:
		return nil, fmt.Errorf("unknown operator: -%T", right)
	}
}

func evalInfixExpression(operator string, left, right any) (any, error) {
	switch {
	case isNumber(left) && isNumber(right):
		return evalNumberInfixExpression(operator, left, right)
	case isString(left) && isString(right):
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return left == right, nil
	case operator == "!=":
		return left != right, nil
	case operator == "&&":
		return isTruthy(left) && isTruthy(right), nil
	case operator == "||":
		return isTruthy(left) || isTruthy(right), nil
	default:
		return nil, fmt.Errorf("unknown operator: %T %s %T", left, operator, right)
	}
}

func isNumber(val any) bool {
	switch val.(type) {
	case int, int64, float64:
		return true
	default:
		return false
	}
}

func isString(val any) bool {
	_, ok := val.(string)
	return ok
}

func toFloat(val any) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}

func evalNumberInfixExpression(operator string, left, right any) (any, error) {
	leftVal := toFloat(left)
	rightVal := toFloat(right)

	switch operator {
	case "+":
		return leftVal + rightVal, nil
	case "-":
		return leftVal - rightVal, nil
	case "*":
		return leftVal * rightVal, nil
	case "/":
		return leftVal / rightVal, nil
	case "<":
		return leftVal < rightVal, nil
	case ">":
		return leftVal > rightVal, nil
	case "==":
		return leftVal == rightVal, nil
	case "!=":
		return leftVal != rightVal, nil
	default:
		return nil, fmt.Errorf("unknown operator: %T %s %T", left, operator, right)
	}
}

func evalStringInfixExpression(operator string, left, right any) (any, error) {
	leftVal := left.(string)
	rightVal := right.(string)
	if operator == "+" {
		return leftVal + rightVal, nil
	}
	if operator == "==" {
		return leftVal == rightVal, nil
	}
	if operator == "!=" {
		return leftVal != rightVal, nil
	}
	return nil, fmt.Errorf("unknown operator: %s %s %s", leftVal, operator, rightVal)
}

func isTruthy(obj any) bool {
	switch obj {
	case nil:
		return false
	case true:
		return true
	case false:
		return false
	default:
		return true
	}
}

func evalIndexExpression(left, index any) (any, error) {
	switch leftVal := left.(type) {
	case map[string]any:
		return evalMapIndexExpression(leftVal, index)
	case []any:
		return evalArrayIndexExpression(leftVal, index)
	default:
		return nil, fmt.Errorf("index operator not supported: %T", left)
	}
}

func evalMapIndexExpression(left map[string]any, index any) (any, error) {
	key, ok := index.(string)
	if !ok {
		return nil, fmt.Errorf("map index must be string, got %T", index)
	}
	if val, ok := left[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func evalArrayIndexExpression(left []any, index any) (any, error) {
	idx, ok := index.(int64)
	if !ok {
		// Try int
		if i, ok := index.(int); ok {
			idx = int64(i)
		} else {
			return nil, fmt.Errorf("array index must be integer, got %T", index)
		}
	}
	if idx < 0 || idx >= int64(len(left)) {
		return nil, fmt.Errorf("index out of bounds: %d", idx)
	}
	return left[idx], nil
}
