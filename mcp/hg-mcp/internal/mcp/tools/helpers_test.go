package tools

import (
	"errors"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// --- helper to build CallToolRequest with arguments ---

func makeReq(args map[string]interface{}) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

func makeReqNilArgs() mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = nil
	return req
}

// extractText returns the text from the first TextContent in a result.
func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("first content is not TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

// ==================== TextResult ====================

func TestTextResult(t *testing.T) {
	r := TextResult("hello world")
	got := extractText(t, r)
	if got != "hello world" {
		t.Errorf("TextResult text = %q, want %q", got, "hello world")
	}
	if r.IsError {
		t.Error("TextResult should not be an error")
	}
}

func TestTextResult_Empty(t *testing.T) {
	r := TextResult("")
	got := extractText(t, r)
	if got != "" {
		t.Errorf("TextResult text = %q, want empty", got)
	}
}

// ==================== ErrorResult ====================

func TestErrorResult(t *testing.T) {
	r := ErrorResult(errors.New("something broke"))
	got := extractText(t, r)
	if got != "something broke" {
		t.Errorf("ErrorResult text = %q, want %q", got, "something broke")
	}
	if !r.IsError {
		t.Error("ErrorResult should be an error")
	}
}

// ==================== JSONResult ====================

func TestJSONResult_Map(t *testing.T) {
	r := JSONResult(map[string]interface{}{"status": "ok", "count": 42})
	got := extractText(t, r)
	if !strings.Contains(got, `"status": "ok"`) {
		t.Errorf("JSONResult missing status field, got: %s", got)
	}
	if !strings.Contains(got, `"count": 42`) {
		t.Errorf("JSONResult missing count field, got: %s", got)
	}
	if r.IsError {
		t.Error("JSONResult should not be an error")
	}
}

func TestJSONResult_Slice(t *testing.T) {
	r := JSONResult([]string{"a", "b"})
	got := extractText(t, r)
	if !strings.Contains(got, `"a"`) || !strings.Contains(got, `"b"`) {
		t.Errorf("JSONResult missing slice elements, got: %s", got)
	}
}

func TestJSONResult_Unmarshalable(t *testing.T) {
	// channels cannot be marshaled to JSON
	r := JSONResult(make(chan int))
	if !r.IsError {
		t.Error("JSONResult with unmarshalable data should be an error")
	}
}

// ==================== GetStringParam ====================

func TestGetStringParam(t *testing.T) {
	req := makeReq(map[string]interface{}{"name": "alice"})
	if got := GetStringParam(req, "name"); got != "alice" {
		t.Errorf("GetStringParam = %q, want %q", got, "alice")
	}
}

func TestGetStringParam_Missing(t *testing.T) {
	req := makeReq(map[string]interface{}{"other": "val"})
	if got := GetStringParam(req, "name"); got != "" {
		t.Errorf("GetStringParam missing key = %q, want empty", got)
	}
}

func TestGetStringParam_WrongType(t *testing.T) {
	req := makeReq(map[string]interface{}{"name": 123})
	if got := GetStringParam(req, "name"); got != "" {
		t.Errorf("GetStringParam wrong type = %q, want empty", got)
	}
}

func TestGetStringParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	if got := GetStringParam(req, "name"); got != "" {
		t.Errorf("GetStringParam nil args = %q, want empty", got)
	}
}

// ==================== GetIntParam ====================

func TestGetIntParam(t *testing.T) {
	req := makeReq(map[string]interface{}{"count": float64(7)})
	if got := GetIntParam(req, "count", 0); got != 7 {
		t.Errorf("GetIntParam = %d, want 7", got)
	}
}

func TestGetIntParam_Default(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	if got := GetIntParam(req, "count", 42); got != 42 {
		t.Errorf("GetIntParam default = %d, want 42", got)
	}
}

func TestGetIntParam_WrongType(t *testing.T) {
	req := makeReq(map[string]interface{}{"count": "not a number"})
	if got := GetIntParam(req, "count", 99); got != 99 {
		t.Errorf("GetIntParam wrong type = %d, want 99", got)
	}
}

func TestGetIntParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	if got := GetIntParam(req, "count", 5); got != 5 {
		t.Errorf("GetIntParam nil args = %d, want 5", got)
	}
}

// ==================== GetBoolParam ====================

func TestGetBoolParam(t *testing.T) {
	req := makeReq(map[string]interface{}{"flag": true})
	if got := GetBoolParam(req, "flag", false); got != true {
		t.Error("GetBoolParam = false, want true")
	}
}

func TestGetBoolParam_Default(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	if got := GetBoolParam(req, "flag", true); got != true {
		t.Error("GetBoolParam default = false, want true")
	}
}

func TestGetBoolParam_WrongType(t *testing.T) {
	req := makeReq(map[string]interface{}{"flag": "yes"})
	if got := GetBoolParam(req, "flag", false); got != false {
		t.Error("GetBoolParam wrong type = true, want false")
	}
}

func TestGetBoolParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	if got := GetBoolParam(req, "flag", true); got != true {
		t.Error("GetBoolParam nil args = false, want true")
	}
}

// ==================== GetFloatParam ====================

func TestGetFloatParam(t *testing.T) {
	req := makeReq(map[string]interface{}{"price": float64(19.99)})
	if got := GetFloatParam(req, "price", 0); got != 19.99 {
		t.Errorf("GetFloatParam = %f, want 19.99", got)
	}
}

func TestGetFloatParam_Default(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	if got := GetFloatParam(req, "price", 9.99); got != 9.99 {
		t.Errorf("GetFloatParam default = %f, want 9.99", got)
	}
}

func TestGetFloatParam_WrongType(t *testing.T) {
	req := makeReq(map[string]interface{}{"price": "expensive"})
	if got := GetFloatParam(req, "price", 1.0); got != 1.0 {
		t.Errorf("GetFloatParam wrong type = %f, want 1.0", got)
	}
}

func TestGetFloatParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	if got := GetFloatParam(req, "price", 3.14); got != 3.14 {
		t.Errorf("GetFloatParam nil args = %f, want 3.14", got)
	}
}

// ==================== HasParam ====================

func TestHasParam_Present(t *testing.T) {
	req := makeReq(map[string]interface{}{"key": "value"})
	if !HasParam(req, "key") {
		t.Error("HasParam should return true for present key")
	}
}

func TestHasParam_Absent(t *testing.T) {
	req := makeReq(map[string]interface{}{"other": "value"})
	if HasParam(req, "key") {
		t.Error("HasParam should return false for absent key")
	}
}

func TestHasParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	if HasParam(req, "key") {
		t.Error("HasParam should return false for nil args")
	}
}

func TestHasParam_NilValue(t *testing.T) {
	// Key exists but value is nil — still counts as present
	req := makeReq(map[string]interface{}{"key": nil})
	if !HasParam(req, "key") {
		t.Error("HasParam should return true when key exists with nil value")
	}
}

// ==================== GetStringArrayParam ====================

func TestGetStringArrayParam(t *testing.T) {
	req := makeReq(map[string]interface{}{
		"tags": []interface{}{"a", "b", "c"},
	})
	got := GetStringArrayParam(req, "tags")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("GetStringArrayParam = %v, want [a b c]", got)
	}
}

func TestGetStringArrayParam_Empty(t *testing.T) {
	req := makeReq(map[string]interface{}{
		"tags": []interface{}{},
	})
	got := GetStringArrayParam(req, "tags")
	if len(got) != 0 {
		t.Errorf("GetStringArrayParam empty = %v, want []", got)
	}
}

func TestGetStringArrayParam_Missing(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	got := GetStringArrayParam(req, "tags")
	if got != nil {
		t.Errorf("GetStringArrayParam missing = %v, want nil", got)
	}
}

func TestGetStringArrayParam_WrongType(t *testing.T) {
	req := makeReq(map[string]interface{}{"tags": "not-an-array"})
	got := GetStringArrayParam(req, "tags")
	if got != nil {
		t.Errorf("GetStringArrayParam wrong type = %v, want nil", got)
	}
}

func TestGetStringArrayParam_MixedTypes(t *testing.T) {
	// Non-string elements should be skipped
	req := makeReq(map[string]interface{}{
		"tags": []interface{}{"a", 42, "b", true},
	})
	got := GetStringArrayParam(req, "tags")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("GetStringArrayParam mixed = %v, want [a b]", got)
	}
}

func TestGetStringArrayParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	got := GetStringArrayParam(req, "tags")
	if got != nil {
		t.Errorf("GetStringArrayParam nil args = %v, want nil", got)
	}
}

// ==================== CodedErrorResult ====================

func TestCodedErrorResult(t *testing.T) {
	r := CodedErrorResult(ErrNotFound, errors.New("item 123 not found"))
	got := extractText(t, r)
	if !strings.HasPrefix(got, "[NOT_FOUND]") {
		t.Errorf("CodedErrorResult prefix missing, got: %s", got)
	}
	if !strings.Contains(got, "item 123 not found") {
		t.Errorf("CodedErrorResult message missing, got: %s", got)
	}
	if !r.IsError {
		t.Error("CodedErrorResult should be an error")
	}
}

func TestCodedErrorResult_AllCodes(t *testing.T) {
	codes := []string{ErrClientInit, ErrInvalidParam, ErrTimeout, ErrNotFound, ErrAPIError, ErrPermission}
	for _, code := range codes {
		r := CodedErrorResult(code, errors.New("test"))
		got := extractText(t, r)
		expected := "[" + code + "] test"
		if got != expected {
			t.Errorf("CodedErrorResult(%s) = %q, want %q", code, got, expected)
		}
	}
}

// ==================== ActionableErrorResult ====================

func TestActionableErrorResult_NoSuggestions(t *testing.T) {
	r := ActionableErrorResult(ErrAPIError, errors.New("connection refused"))
	got := extractText(t, r)
	if !strings.HasPrefix(got, "[API_ERROR]") {
		t.Errorf("ActionableErrorResult prefix missing, got: %s", got)
	}
	if strings.Contains(got, "Suggested actions") {
		t.Error("ActionableErrorResult with no suggestions should not contain 'Suggested actions'")
	}
	if !r.IsError {
		t.Error("ActionableErrorResult should be an error")
	}
}

func TestActionableErrorResult_WithSuggestions(t *testing.T) {
	r := ActionableErrorResult(
		ErrClientInit,
		errors.New("cannot connect"),
		"Check that the service is running",
		"Verify credentials in .env",
	)
	got := extractText(t, r)
	if !strings.Contains(got, "Suggested actions:") {
		t.Errorf("ActionableErrorResult missing suggestions header, got: %s", got)
	}
	if !strings.Contains(got, "Check that the service is running") {
		t.Errorf("ActionableErrorResult missing first suggestion, got: %s", got)
	}
	if !strings.Contains(got, "Verify credentials in .env") {
		t.Errorf("ActionableErrorResult missing second suggestion, got: %s", got)
	}
}

// ==================== Edge cases: non-map arguments ====================

func TestGetters_NonMapArguments(t *testing.T) {
	// Arguments set to something other than map[string]interface{}
	req := mcp.CallToolRequest{}
	req.Params.Arguments = "not a map"

	if got := GetStringParam(req, "x"); got != "" {
		t.Errorf("GetStringParam non-map = %q, want empty", got)
	}
	if got := GetIntParam(req, "x", 10); got != 10 {
		t.Errorf("GetIntParam non-map = %d, want 10", got)
	}
	if got := GetBoolParam(req, "x", true); got != true {
		t.Error("GetBoolParam non-map = false, want true")
	}
	if got := GetFloatParam(req, "x", 2.5); got != 2.5 {
		t.Errorf("GetFloatParam non-map = %f, want 2.5", got)
	}
	if HasParam(req, "x") {
		t.Error("HasParam non-map should return false")
	}
	if got := GetStringArrayParam(req, "x"); got != nil {
		t.Errorf("GetStringArrayParam non-map = %v, want nil", got)
	}
}

// ==================== RequireStringParam ====================

func TestRequireStringParam_Valid(t *testing.T) {
	req := makeReq(map[string]interface{}{"name": "alice"})
	val, errResult := RequireStringParam(req, "name")
	if errResult != nil {
		t.Fatal("RequireStringParam returned error for valid param")
	}
	if val != "alice" {
		t.Errorf("RequireStringParam = %q, want %q", val, "alice")
	}
}

func TestRequireStringParam_Missing(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	val, errResult := RequireStringParam(req, "name")
	if errResult == nil {
		t.Fatal("RequireStringParam should return error for missing param")
	}
	if val != "" {
		t.Errorf("RequireStringParam val = %q, want empty on error", val)
	}
	if !errResult.IsError {
		t.Error("RequireStringParam error result should have IsError=true")
	}
	got := extractText(t, errResult)
	if !strings.Contains(got, "name") || !strings.Contains(got, "required") {
		t.Errorf("RequireStringParam error message = %q, want mention of 'name' and 'required'", got)
	}
}

func TestRequireStringParam_Empty(t *testing.T) {
	req := makeReq(map[string]interface{}{"name": ""})
	_, errResult := RequireStringParam(req, "name")
	if errResult == nil {
		t.Fatal("RequireStringParam should return error for empty string")
	}
}

func TestRequireStringParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	_, errResult := RequireStringParam(req, "name")
	if errResult == nil {
		t.Fatal("RequireStringParam should return error for nil args")
	}
}

// ==================== RequireIntParam ====================

func TestRequireIntParam_Valid(t *testing.T) {
	req := makeReq(map[string]interface{}{"count": float64(42)})
	val, errResult := RequireIntParam(req, "count")
	if errResult != nil {
		t.Fatal("RequireIntParam returned error for valid param")
	}
	if val != 42 {
		t.Errorf("RequireIntParam = %d, want 42", val)
	}
}

func TestRequireIntParam_Missing(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	val, errResult := RequireIntParam(req, "count")
	if errResult == nil {
		t.Fatal("RequireIntParam should return error for missing param")
	}
	if val != 0 {
		t.Errorf("RequireIntParam val = %d, want 0 on error", val)
	}
	if !errResult.IsError {
		t.Error("RequireIntParam error result should have IsError=true")
	}
}

func TestRequireIntParam_Zero(t *testing.T) {
	req := makeReq(map[string]interface{}{"count": float64(0)})
	_, errResult := RequireIntParam(req, "count")
	if errResult == nil {
		t.Fatal("RequireIntParam should return error for zero value")
	}
}

func TestRequireIntParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	_, errResult := RequireIntParam(req, "count")
	if errResult == nil {
		t.Fatal("RequireIntParam should return error for nil args")
	}
}

// ==================== RequireFloatParam ====================

func TestRequireFloatParam_Valid(t *testing.T) {
	req := makeReq(map[string]interface{}{"price": float64(19.99)})
	val, errResult := RequireFloatParam(req, "price")
	if errResult != nil {
		t.Fatal("RequireFloatParam returned error for valid param")
	}
	if val != 19.99 {
		t.Errorf("RequireFloatParam = %f, want 19.99", val)
	}
}

func TestRequireFloatParam_Missing(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	val, errResult := RequireFloatParam(req, "price")
	if errResult == nil {
		t.Fatal("RequireFloatParam should return error for missing param")
	}
	if val != 0 {
		t.Errorf("RequireFloatParam val = %f, want 0 on error", val)
	}
}

func TestRequireFloatParam_Zero(t *testing.T) {
	req := makeReq(map[string]interface{}{"price": float64(0)})
	_, errResult := RequireFloatParam(req, "price")
	if errResult == nil {
		t.Fatal("RequireFloatParam should return error for zero value")
	}
}

func TestRequireFloatParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	_, errResult := RequireFloatParam(req, "price")
	if errResult == nil {
		t.Fatal("RequireFloatParam should return error for nil args")
	}
}

// ==================== RequireStringArrayParam ====================

func TestRequireStringArrayParam_Valid(t *testing.T) {
	req := makeReq(map[string]interface{}{
		"tags": []interface{}{"a", "b"},
	})
	val, errResult := RequireStringArrayParam(req, "tags")
	if errResult != nil {
		t.Fatal("RequireStringArrayParam returned error for valid param")
	}
	if len(val) != 2 || val[0] != "a" || val[1] != "b" {
		t.Errorf("RequireStringArrayParam = %v, want [a b]", val)
	}
}

func TestRequireStringArrayParam_Empty(t *testing.T) {
	req := makeReq(map[string]interface{}{
		"tags": []interface{}{},
	})
	_, errResult := RequireStringArrayParam(req, "tags")
	if errResult == nil {
		t.Fatal("RequireStringArrayParam should return error for empty array")
	}
}

func TestRequireStringArrayParam_Missing(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	_, errResult := RequireStringArrayParam(req, "tags")
	if errResult == nil {
		t.Fatal("RequireStringArrayParam should return error for missing param")
	}
}

func TestRequireStringArrayParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	_, errResult := RequireStringArrayParam(req, "tags")
	if errResult == nil {
		t.Fatal("RequireStringArrayParam should return error for nil args")
	}
}

// ==================== OptionalStringParam ====================

func TestOptionalStringParam_Present(t *testing.T) {
	req := makeReq(map[string]interface{}{"mode": "dark"})
	got := OptionalStringParam(req, "mode", "light")
	if got != "dark" {
		t.Errorf("OptionalStringParam = %q, want %q", got, "dark")
	}
}

func TestOptionalStringParam_Missing(t *testing.T) {
	req := makeReq(map[string]interface{}{})
	got := OptionalStringParam(req, "mode", "light")
	if got != "light" {
		t.Errorf("OptionalStringParam missing = %q, want %q", got, "light")
	}
}

func TestOptionalStringParam_Empty(t *testing.T) {
	req := makeReq(map[string]interface{}{"mode": ""})
	got := OptionalStringParam(req, "mode", "light")
	if got != "light" {
		t.Errorf("OptionalStringParam empty = %q, want %q (default)", got, "light")
	}
}

func TestOptionalStringParam_NilArgs(t *testing.T) {
	req := makeReqNilArgs()
	got := OptionalStringParam(req, "mode", "auto")
	if got != "auto" {
		t.Errorf("OptionalStringParam nil args = %q, want %q", got, "auto")
	}
}

// ==================== truncateResponse ====================

func TestTruncateResponse_Nil(t *testing.T) {
	if got := truncateResponse(nil); got != nil {
		t.Error("truncateResponse(nil) should return nil")
	}
}

func TestTruncateResponse_Small(t *testing.T) {
	r := TextResult("small text")
	truncated := truncateResponse(r)
	got := extractText(t, truncated)
	if got != "small text" {
		t.Errorf("truncateResponse small = %q, want %q", got, "small text")
	}
}

func TestTruncateResponse_Large(t *testing.T) {
	large := strings.Repeat("x", MaxResponseSize+100)
	r := TextResult(large)
	truncated := truncateResponse(r)
	got := extractText(t, truncated)
	if !strings.HasSuffix(got, "[TRUNCATED: response exceeded 128KB limit]") {
		t.Error("truncateResponse should add truncation suffix")
	}
	if len(got) < MaxResponseSize {
		t.Error("truncateResponse should preserve content up to MaxResponseSize")
	}
}
