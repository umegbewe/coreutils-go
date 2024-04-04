package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: expr <expression>")
		fmt.Println("Supported operations: +, -, *, /, %, ^, ==, !=, <, <=, >, >=, &&, ||, &, |, ^^, <<, >>")
		os.Exit(1)
	}

	expression := os.Args[1]
	tokens := tokenize(expression)
	postfix, err := shuntingYard(tokens)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	result, err := evaluatePostfix(postfix)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

func tokenize(expression string) []string {
	re := regexp.MustCompile(`(\d+|==|!=|<=|>=|&&|\|\||<<|>>|[+\-*/^&|%<>()])`)
	return re.FindAllString(expression, -1)
}

func shuntingYard(tokens []string) ([]string, error) {
    var queue []string
    var operatorStack []string

    precedence := map[string]int{
        "+": 2, "-": 2, "*": 3, "/": 3, "%": 3,
        "^": 4, "==": 1, "!=": 1, "<": 1, "<=": 1,
        ">": 1, ">=": 1, "&&": 0, "||": 0,
        "&": 3, "|": 2, "^^": 3, "<<": 3, ">>": 3,
    }

    associativity := map[string]string{
        "+": "left", "-": "left", "*": "left", "/": "left", "%": "left",
        "^": "right", "==": "left", "!=": "left", "<": "left", "<=": "left",
        ">": "left", ">=": "left", "&&": "left", "||": "left",
        "&": "left", "|": "left", "^^": "left", "<<": "left", ">>": "left",
    }

    for _, token := range tokens {
        switch {
        case token == "(":
            operatorStack = append(operatorStack, token)
        case token == ")":
            for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" {
                queue = append(queue, operatorStack[len(operatorStack)-1])
                operatorStack = operatorStack[:len(operatorStack)-1]
            }
            if len(operatorStack) == 0 {
                return nil, fmt.Errorf("mismatched parentheses")
            }
            operatorStack = operatorStack[:len(operatorStack)-1] // pop "("
        case token == "&&" || token == "||":
            for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" &&
                precedence[token] <= precedence[operatorStack[len(operatorStack)-1]] {
                queue = append(queue, operatorStack[len(operatorStack)-1])
                operatorStack = operatorStack[:len(operatorStack)-1]
            }
            operatorStack = append(operatorStack, token)
        case precedence[token] > 0:
            for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" &&
                ((associativity[token] == "left" && precedence[token] <= precedence[operatorStack[len(operatorStack)-1]]) ||
                    (associativity[token] == "right" && precedence[token] < precedence[operatorStack[len(operatorStack)-1]])) {
                queue = append(queue, operatorStack[len(operatorStack)-1])
                operatorStack = operatorStack[:len(operatorStack)-1]
            }
            operatorStack = append(operatorStack, token)
        default:
            queue = append(queue, token)
        }
    }

    // pop any remaining operators
    for len(operatorStack) > 0 {
        if operatorStack[len(operatorStack)-1] == "(" {
            return nil, fmt.Errorf("mismatched parentheses")
        }
        queue = append(queue, operatorStack[len(operatorStack)-1])
        operatorStack = operatorStack[:len(operatorStack)-1]
    }

    return queue, nil
}
func evaluatePostfix(postfix []string) (interface{}, error) {
	var stack []interface{}
	for _, token := range postfix {
		if num, err := strconv.Atoi(token); err == nil {
			stack = append(stack, num)
		} else if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				return nil, fmt.Errorf("invalid expression")
			}
			b, a := stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			var result interface{}
			switch token {
			case "+":
				result = add(a, b)
			case "-":
				result = subtract(a, b)
			case "*":
				result = multiply(a, b)
			case "/":
				if b == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = divide(a, b)
			case "%":
				if isFloat(a) || isFloat(b) {
					return nil, fmt.Errorf("unsupported operation: %% on floating-point numbers")
				}
				if b.(int) == 0 {
					return nil, fmt.Errorf("modulo by zero")
				}
				result = modulo(a.(int), b.(int))
			case "^":
				result = power(a, b)
			case "==":
				result = equals(a, b)
			case "!=":
				result = notEquals(a, b)
			case "<":
				result = lessThan(a, b)
			case "<=":
				result = lessThanOrEqual(a, b)
			case ">":
				result = greaterThan(a, b)
			case ">=":
				result = greaterThanOrEqual(a, b)
			case "&&":
				result = and(a, b)
			case "||":
				result = or(a, b)
			case "&":
                if isFloat(a) || isFloat(b) {
                    return nil, fmt.Errorf("unsupported operation: & on floating-point numbers")
                }
                result = bitwiseAnd(a.(int), b.(int))
            case "|":
                if isFloat(a) || isFloat(b) {
                    return nil, fmt.Errorf("unsupported operation: | on floating-point numbers")
                }
                result = bitwiseOr(a.(int), b.(int))
            case "^^":
                if isFloat(a) || isFloat(b) {
                    return nil, fmt.Errorf("unsupported operation: ^^ on floating-point numbers")
                }
                result = bitwiseXor(a.(int), b.(int))
            case "<<":
                if isFloat(a) || isFloat(b) {
                    return nil, fmt.Errorf("unsupported operation: << on floating-point numbers")
                }
                result = leftShift(a.(int), b.(int))
            case ">>":
                if isFloat(a) || isFloat(b) {
                    return nil, fmt.Errorf("unsupported operation: >> on floating-point numbers")
                }
                result = rightShift(a.(int), b.(int))
			default:
				return nil, fmt.Errorf("unsupported operation: %s", token)
			}
			stack = append(stack, result)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid expression")
	}

	return stack[0], nil
}
