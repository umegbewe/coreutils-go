package main

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		expression string
		expected []string
	}{
		{"3 + 4", []string{"3", "+", "4"}},
		{"3 + * 4", []string{"3", "+", "*", "4"}},
		{"3 + 4 * 2 / ( 1 - 5 ) ^ 2 ^ 3", []string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5", ")", "^", "2", "^", "3"}},
		{"3 + 4 * 2 / ( 1 - 5 ) ^ 2 ^ 3", []string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5", ")", "^", "2", "^", "3"}},
		{"(2+2)*3", []string{"(", "2", "+", "2", ")", "*", "3"}},
        {"2 + 2 * 3", []string{"2", "+", "2", "*", "3"}},
		{"2 == 2", []string{"2", "==", "2"}},
        {"2 && 3", []string{"2", "&&", "3"}},
        {"2 << 3", []string{"2", "<<", "3"}},
	}

	for _, test := range tests {
		if got := tokenize(test.expression); !reflect.DeepEqual(got, test.expected) {
			t.Errorf("expected %v, got %v", test.expected, got)
		}
	}
}

func TestShuntingYard(t *testing.T) {
	tests := []struct {
		tokens []string
		expected []string
	}{
		{[]string{"3", "+", "4"}, []string{"3", "4", "+"}},
		{[]string{"(", "3", "+", "4", ")", "*", "2"}, []string{"3", "4", "+", "2", "*"}},
		{[]string{"3", "*", "(","4", "+", "5", ")"}, []string{"3", "4", "5", "+", "*"}},
		{[]string{"2", "==", "2"}, []string{"2", "2", "=="}},
        {[]string{"2", "&&", "3"}, []string{"2", "3", "&&"}},
        {[]string{"2", "<<", "3"}, []string{"2", "3", "<<"}},
        {[]string{"2", "*", "(", "3", "+", "4", ")"}, []string{"2", "3", "4", "+", "*"}},
		{[]string{"(", "2", "+", "3", ")", "*", "(", "4", "-", "1", ")"}, []string{"2", "3", "+", "4", "1", "-", "*"}},
        {[]string{"2", "*", "3", "+", "4"}, []string{"2", "3", "*", "4", "+"}},
        {[]string{"2", "+", "3", "*", "4"}, []string{"2", "3", "4", "*", "+"}},
        {[]string{"2", "*", "(", "3", "+", "4", ")"}, []string{"2", "3", "4", "+", "*"}},
        {[]string{"(", "2", "*", "3", ")", "+", "4"}, []string{"2", "3", "*", "4", "+"}},
        {[]string{"2", "<<", "3", ">>", "1"}, []string{"2", "3", "<<", "1", ">>"}},
        {[]string{"2", "&&", "3", "||", "4"}, []string{"2", "3", "&&", "4", "||"}},
        {[]string{"2", "+", "2", "*", "3", "-", "4", "/", "2"}, []string{"2", "2", "3", "*", "+", "4", "2", "/", "-"}},
	}

	for _, test := range tests {
		if got, _ := shuntingYard(test.tokens); !reflect.DeepEqual(got, test.expected) {
			t.Errorf("expected %v, got %v", test.expected, got)
		}
	}
}

func TestShuntingYardEdgeCase(t *testing.T) {
	
	tokens := []string{"3", "+", "(", "4", "*", "5"}
	
	_, err := shuntingYard(tokens)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestEvaluatePostfix(t *testing.T) {
    tests := []struct {
        postfix    []string
        expected   interface{}
        expectError bool
    }{
        {[]string{"3", "4", "+"}, 7, false},
        {[]string{"3", "4", "+", "2", "*"}, 14, false},
        {[]string{"3", "4", "5", "+", "*"}, 27, false},
        {[]string{"2", "2", "=="}, 1, false},
        {[]string{"2", "3", "&&"}, 1, false}, 
        {[]string{"2", "3", "<<"}, 16, false},
        {[]string{"10", "2", "/"}, 5, false},
        {[]string{"10", "0", "/"}, nil, true},
        {[]string{"2", "3", "+", "4", "1", "-", "*"}, 15, false},
        {[]string{"2", "3", "*", "4", "+"}, 10, false},
        {[]string{"2", "3", "4", "*", "+"}, 14, false},
        {[]string{"2", "3", "4", "+", "*"}, 14, false},
        {[]string{"2", "3", "*", "4", "+"}, 10, false},
        {[]string{"2", "3", "<<", "1", ">>"}, 8, false},
        {[]string{"2", "3", "&&", "4", "||"}, 1, false},
        {[]string{"2", "2", "3", "*", "+", "4", "2", "/", "-"}, 6, false},
        {[]string{"2", "0", "%"}, nil, true},
    }
    for _, test := range tests {
        got, err := evaluatePostfix(test.postfix)
        if test.expectError {
            if err == nil {
                t.Errorf("expected an error, but got none")
            }
        } else {
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if !reflect.DeepEqual(got, test.expected) {
                t.Errorf("for postfix %v, expected %v, got %v", test.postfix, test.expected, got)
            }
        }
    }
}


func TestEvaluatePostfixEdgeCase(t *testing.T) {
	postfix := []string{"3", "4", "5", "+", "+", "+"}

	_, err := evaluatePostfix(postfix)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
