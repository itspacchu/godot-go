package extensionapi

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/iancoleman/strcase"
)

var (
	goReturnType = goArgumentType
)

func goFormatFieldName(n string) string {
	n = goArgumentName(n)

	return fmt.Sprintf("%s%s", strings.ToUpper(n[:1]), n[1:])
}

var (
	underscoreDigitRe = regexp.MustCompile(`_(\d)`)
	digitDRe          = regexp.MustCompile(`(\d)_([iIdD])`)
)

func screamingSnake(v string) string {
	v = strcase.ToScreamingSnake(v)

	return digitDRe.ReplaceAllString(underscoreDigitRe.ReplaceAllString(v, `$1`), `${1}${2}`)
}

func goVariantConstructor(t, innerText string) string {
	switch t {
	case "float", "real_t", "double":
		return fmt.Sprintf("NewVariantFloat64(%s)", innerText)
	case "int", "uint64_t":
		return fmt.Sprintf("NewVariantInt64(%s)", innerText)
	case "bool":
		return fmt.Sprintf("NewVariantBool(%s)", innerText)
	case "String":
		return fmt.Sprintf("NewVariantString(%s)", innerText)
	case "StringName":
		return fmt.Sprintf("NewVariantStringName(%s)", innerText)
	default:
		return fmt.Sprintf("NewVariantWrapped(&%s)", innerText)
	}
}

func goArgumentName(t string) string {
	switch t {
	case "string":
		return "strValue"
	case "internal":
		return "internalMode"
	case "type":
		return "typeName"
	case "range":
		return "valueRange"
	case "default":
		return "defaultName"
	case "interface":
		return "interfaceName"
	case "map":
		return "resourceMap"
	case "var":
		return "varName"
	case "func":
		return "callbackFunc"
	default:
		return t
	}
}

func typeHasPtr(t string) bool {
	switch t {
	case "float", "int", "Object":
		return false
	default:
		return true
	}
}

func goArgumentType(t string) string {
	if strings.HasPrefix(t, "enum::") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "const ") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "bitfield") {
		t = t[8:]
	}

	var (
		indirection int
	)

	if strings.HasSuffix(t, "**") {
		indirection = 2
		t = strings.TrimSpace(t[:len(t)-2])
	}

	if strings.HasSuffix(t, "*") {
		indirection = 1
		t = strings.TrimSpace(t[:len(t)-1])
	}

	switch t {
	case "void":
		switch indirection {
		case 0:
			return ""
		case 1:
			return "unsafe.Pointer"
		case 2:
			return "*unsafe.Pointer"
		default:
			panic("unexepected pointer indirection")
		}
	case "Vector2i", "Vector3i", "Vector4i", "Rect2i":
	case "float", "real_t":
		t = "float32"
	case "double":
		t = "float64"
	case "int":
		t = "int64"
	case "uint64_t":
		t = "uint64"
	case "bool":
		t = "bool"
	case "String":
		t = "String"
	case "Nil":
		t = "Variant"
	case "":
		t = ""
	default:
		t = strcase.ToCamel(t)
	}

	return strings.Repeat("*", indirection) + t
}

func goHasArgumentTypeEncoder(t string) bool {
	if strings.HasPrefix(t, "enum::") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "const ") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "bitfield") {
		t = t[8:]
	}

	var (
		indirection int
	)

	if strings.HasSuffix(t, "**") {
		indirection = 2
		t = strings.TrimSpace(t[:len(t)-2])
	}

	if strings.HasSuffix(t, "*") {
		indirection = 1
		t = strings.TrimSpace(t[:len(t)-1])
	}

	switch t {
	case "void":
		switch indirection {
		case 0:
			return false
		case 1:
			return false
		case 2:
			return false
		default:
			panic("unexepected pointer indirection")
		}
	case "Vector2i", "Vector3i", "Vector4i", "Rect2i":
		return true
	case "float", "real_t":
		return true
	case "double":
		return true
	case "int":
		return true
	case "uint64_t":
		return true
	case "bool":
		return true
	case "String":
		return true
	case "Nil":
		return true
	case "":
		return false
	}

	return true
}

// func goClassImpl(c string) string {
// 	return strcase.ToLowerCamel(c)
// }

func goClassEnumName(c, e, n string) string {
	return fmt.Sprintf("%s_%s_%s",
		strings.ToUpper(strcase.ToSnake(c)),
		strings.ToUpper(strcase.ToSnake(e)),
		strings.ToUpper(strcase.ToSnake(n)))
}

func goMethodName(n string) string {
	if strings.HasPrefix(n, "_") {
		return fmt.Sprintf("Internal_%s", strcase.ToCamel(n))
	}

	return strcase.ToCamel(n)
}

func nativeStructureFormatToFields(f string) string {
	sb := strings.Builder{}
	fields := strings.Split(f, ";")

	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
		pair := strings.SplitN(fields[i], " ", 2)

		t := strings.TrimSpace(pair[0])
		n := strings.TrimSpace(pair[1])

		if strings.Contains(n, "=") {
			nPairs := strings.SplitN(n, "=", 2)
			n = nPairs[0]
		}

		if strings.HasPrefix(n, "(*") {
			sb.WriteString("/* ")

			hasPointer := strings.HasPrefix(n, "*")

			if hasPointer {
				n = n[1:]
			}

			sb.WriteString(goFormatFieldName(n))
			sb.WriteString(" ")

			if strings.HasPrefix(n, "*") {
				sb.WriteString("*")
			}

			sb.WriteString(goArgumentType(t))
			sb.WriteString("*/\n")
		} else {

			hasPointer := strings.HasPrefix(n, "*")

			if hasPointer {
				n = n[1:]
			}

			sb.WriteString(goFormatFieldName(n))
			sb.WriteString(" ")

			if strings.HasPrefix(n, "*") {
				sb.WriteString("*")
			}

			sb.WriteString(goArgumentType(t))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

var (
	operatorIdName = map[string]string{
		"==":     "equal",
		"!=":     "not_equal",
		"<":      "less",
		"<=":     "less_equal",
		">":      "greater",
		">=":     "greater_equal",
		"+":      "add",
		"-":      "subtract",
		"*":      "multiply",
		"/":      "divide",
		"unary-": "negate",
		"unary+": "positive",
		"%":      "module", // this seems like a mispelling, but it stems from gdnative_interface.h constant GDNATIVE_VARIANT_OP_MODULE
		"<<":     "shift_left",
		">>":     "shift_right",
		"&":      "bit_and",
		"|":      "bit_or",
		"^":      "bit_xor",
		"~":      "bit_negate",
		"and":    "and",
		"or":     "or",
		"xor":    "xor",
		"not":    "not",
		"in":     "in",
	}
)

func getOperatorIdName(op string) string {
	return operatorIdName[op]
}

func lowerFirstChar(n string) string {
	return fmt.Sprintf("%s%s", strings.ToLower(n[:1]), n[1:])
}

func upperFirstChar(n string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(n[:1]), n[1:])
}

var (
	needsCopySet = map[string]struct{}{
		"Dictionary": {},
	}
)

func needsCopyInsteadOfMove(typeName string) bool {
	_, ok := needsCopySet[typeName]

	return ok
}

func isCopyConstructor(typeName string, c extensionapiparser.ClassConstructor) bool {
	return len(c.Arguments) == 1 && c.Arguments[0].Type == typeName
}