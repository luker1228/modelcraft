package bizutils

import (
	"testing"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestStandardJSONParser_Marshal(t *testing.T) {
	parser := NewStandardJSONParser()

	data := TestStruct{
		Name: "test",
		Age:  25,
	}

	result, err := parser.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{"name":"test","age":25}`
	if string(result) != expected {
		t.Errorf("Expected %s, got %s", expected, string(result))
	}
}

func TestStandardJSONParser_Unmarshal(t *testing.T) {
	parser := NewStandardJSONParser()

	jsonData := []byte(`{"name":"test","age":25}`)
	var result TestStruct

	err := parser.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Name != "test" || result.Age != 25 {
		t.Errorf("Expected {name: test, Age: 25}, got %+v", result)
	}
}

func TestStandardJSONParser_MarshalToString(t *testing.T) {
	parser := NewStandardJSONParser()

	data := TestStruct{
		Name: "test",
		Age:  25,
	}

	result, err := parser.MarshalToString(data)
	if err != nil {
		t.Fatalf("MarshalToString failed: %v", err)
	}

	expected := `{"name":"test","age":25}`
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestStandardJSONParser_UnmarshalFromString(t *testing.T) {
	parser := NewStandardJSONParser()

	jsonString := `{"name":"test","age":25}`
	var result TestStruct

	err := parser.UnmarshalFromString(jsonString, &result)
	if err != nil {
		t.Fatalf("UnmarshalFromString failed: %v", err)
	}

	if result.Name != "test" || result.Age != 25 {
		t.Errorf("Expected {name: test, Age: 25}, got %+v", result)
	}
}

func TestDefaultJSONParser(t *testing.T) {
	data := TestStruct{
		Name: "test",
		Age:  25,
	}

	// 测试默认解析器的Marshal
	result, err := Marshal(data)
	if err != nil {
		t.Fatalf("Default Marshal failed: %v", err)
	}

	// 测试默认解析器的Unmarshal
	var parsed TestStruct
	err = Unmarshal(result, &parsed)
	if err != nil {
		t.Fatalf("Default Unmarshal failed: %v", err)
	}

	if parsed.Name != "test" || parsed.Age != 25 {
		t.Errorf("Expected {name: test, Age: 25}, got %+v", parsed)
	}
}

func TestDefaultJSONParserString(t *testing.T) {
	data := TestStruct{
		Name: "test",
		Age:  25,
	}

	// 测试默认解析器的MarshalToString
	result, err := MarshalToString(data)
	if err != nil {
		t.Fatalf("Default MarshalToString failed: %v", err)
	}

	// 测试默认解析器的UnmarshalFromString
	var parsed TestStruct
	err = UnmarshalFromString(result, &parsed)
	if err != nil {
		t.Fatalf("Default UnmarshalFromString failed: %v", err)
	}

	if parsed.Name != "test" || parsed.Age != 25 {
		t.Errorf("Expected {name: test, Age: 25}, got %+v", parsed)
	}
}
