package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
)

type FieldModel struct {
	Name     string
	DataType string
	IsList   bool
}

const (
	Integer = "int"
	Double  = "double"
	String  = "String"
	Boolean = "bool"
	Dynamic = "dynamic"
	Object  = "OBJECT"
	Array   = "ARRAY"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("usage: json2dart [filename]\n\n")
		os.Exit(1)
	}

	jsonFile := readJsonFile(os.Args[1])

	os.Chdir(path.Dir(jsonFile.Name()))

	m := map[string]interface{}{}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("file read all: ", err)
		os.Exit(1)
	}

	err = json.Unmarshal([]byte(byteValue), &m)
	if err != nil {
		fmt.Println("json unmarshall error: ", err)
		os.Exit(1)
	}

	parseMap(getFileNameWithoutExtension(jsonFile.Name()), m)

	defer jsonFile.Close()
}

func parseMap(objectName string, aMap map[string]interface{}) {
	objectFields := make(map[string]FieldModel)

	for key, val := range aMap {
		switch val.(type) {
		case map[string]interface{}:
			parseMap(key, val.(map[string]interface{}))
			objectFields[key] = FieldModel{
				DataType: Object,
				Name:     key,
				IsList:   false,
			}
		case []interface{}:
			arrayTmp := parseArray(key, val.([]interface{}))
			objectFields[key] = FieldModel{
				DataType: arrayTmp.DataType,
				Name:     key,
				IsList:   true,
			}
		case float64:
			v := reflect.ValueOf(val)
			v = reflect.Indirect(v)
			floatType := reflect.TypeOf(float64(0))

			var attributeTmp FieldModel

			fvFloat := v.Convert(floatType)
			if (fvFloat.Float() - float64(int64(fvFloat.Float()))) == 0 {
				attributeTmp = FieldModel{
					DataType: Integer,
					Name:     key,
					IsList:   false,
				}
			} else {
				attributeTmp = FieldModel{
					DataType: Double,
					Name:     key,
					IsList:   false,
				}
			}
			objectFields[key] = attributeTmp
		case string:
			objectFields[key] = FieldModel{
				DataType: String,
				Name:     key,
				IsList:   false,
			}
		case bool:
			objectFields[key] = FieldModel{
				DataType: Boolean,
				Name:     key,
				IsList:   false,
			}
		}
	}

	generateDartModelComponents(objectName, objectFields)
}

func parseArray(key string, anArray []interface{}) FieldModel {
	arrayDataTypes := make(map[string]bool)
	var finalArrayDataType FieldModel

	for _, val := range anArray {
		switch val.(type) {
		case map[string]interface{}:
			parseMap(key, val.(map[string]interface{}))
			return FieldModel{
				DataType: Object,
				Name:     key,
				IsList:   true,
			}
		case []interface{}:
			arrayTmp := parseArray(key+"List", val.([]interface{}))
			return FieldModel{
				DataType: formatString("List<%s>", arrayTmp.DataType),
				Name:     key,
				IsList:   true,
			}
		case float64:
			v := reflect.ValueOf(val)
			v = reflect.Indirect(v)
			floatType := reflect.TypeOf(float64(0))

			var attributeTmp FieldModel

			fvFloat := v.Convert(floatType)
			if (fvFloat.Float() - float64(int64(fvFloat.Float()))) == 0 {
				attributeTmp = FieldModel{
					DataType: Integer,
					Name:     key,
					IsList:   false,
				}
			} else {
				attributeTmp = FieldModel{
					DataType: Double,
					Name:     key,
					IsList:   false,
				}
			}
			arrayDataTypes[attributeTmp.DataType] = true
			finalArrayDataType = attributeTmp
		case int:

			finalArrayDataType = FieldModel{
				DataType: Integer,
				Name:     key,
				IsList:   true,
			}
			arrayDataTypes["Integer"] = true
		case string:
			finalArrayDataType = FieldModel{
				DataType: String,
				Name:     key,
				IsList:   true,
			}
			arrayDataTypes["String"] = true
		case bool:
			finalArrayDataType = FieldModel{
				DataType: Boolean,
				Name:     key,
				IsList:   true,
			}
			arrayDataTypes["Boolean"] = true
		}
	}

	if len(arrayDataTypes) == 1 {
		return finalArrayDataType
	}

	return FieldModel{
		DataType: Dynamic,
		Name:     key,
		IsList:   true,
	}
}

func generateDartModelComponents(objectName string, objectFields map[string]FieldModel) {
	importList := make([]string, 0)
	attributeList := make([]string, 0)
	constructorList := make([]string, 0)
	fromJsonList := make([]string, 0)
	preFromJsonList := make([]string, 0)
	toJsonList := make([]string, 0)
	preToJsonList := make([]string, 0)

	for key, val := range objectFields {
		if val.IsList {
			switch val.DataType {
			case Object:
				modelName := strings.Title(key) + "Model"
				valueName := key + "Tmp"
				listName := key + "List"

				attributeName := toCamelCase(key)
				appendStringSlice(&attributeList, formatString("List<%s> %s;", modelName, attributeName))
				appendStringSlice(&constructorList, formatString("this.%s,", attributeName))

				preFromJsonTmp := formatString("var %s = data[\"%s\"] as List;\nList<%s> %s = %s.map((i) => %s.fromJson(i)).toList();", listName, key, modelName, valueName, listName, modelName)
				appendStringSlice(&preFromJsonList, preFromJsonTmp)
				appendStringSlice(&fromJsonList, formatString("%s : %sTmp,", attributeName, key))

				appendStringSlice(&preToJsonList, formatString("List<Map<String, dynamic>> %s = this.%s.map((i) => i.toJson()).toList();", listName, attributeName))
				appendStringSlice(&toJsonList, formatString("\"%s\" : %s,", key, listName))
				appendStringSlice(&importList, formatString("import \"%s_model.dart\";", key))

			default:
				dataType := formatString("%v", val.DataType)
				attributeName := toCamelCase(key)

				appendStringSlice(&attributeList, formatString("List<%s> %s;", dataType, attributeName))
				appendStringSlice(&constructorList, formatString("this.%s,", attributeName))

				preFromJsonTmp := formatString("List<%s> %sTmp = data['%s'].cast<%s>();", val.DataType, key, key, val.DataType)
				appendStringSlice(&preFromJsonList, preFromJsonTmp)

				appendStringSlice(&fromJsonList, formatString("%s : %sTmp,", attributeName, key))
				appendStringSlice(&toJsonList, formatString("\"%s\" : this.%s,", key, attributeName))
			}
		} else {
			switch val.DataType {
			case Object:
				valueName := key + "Tmp"
				modelName := strings.Title(key) + "Model"
				attributeName := toCamelCase(key)

				appendStringSlice(&attributeList, formatString("%s %s;", modelName, attributeName))
				appendStringSlice(&constructorList, formatString("this.%s,", attributeName))

				appendStringSlice(&preFromJsonList, formatString("%s %s = %s.fromJson(data[\"%s\"]);", modelName, valueName, modelName, key))
				appendStringSlice(&fromJsonList, formatString("%s : %s,", attributeName, valueName))

				appendStringSlice(&preToJsonList, formatString("Map<String, dynamic> %s = this.%s.toJson();", valueName, key))
				appendStringSlice(&toJsonList, formatString("\"%s\" : %s,", key, valueName))

				appendStringSlice(&importList, formatString("import \"%s_model.dart\";", key))
			default:
				dataType := formatString("%v", val.DataType)
				attributeName := toCamelCase(key)

				appendStringSlice(&attributeList, formatString("%s %s;", dataType, attributeName))
				appendStringSlice(&constructorList, formatString("this.%s,", attributeName))

				appendStringSlice(&fromJsonList, formatString("%s : data['%s'],", attributeName, key))

				appendStringSlice(&toJsonList, formatString("\"%s\" : this.%s,", key, attributeName))
			}
		}
	}

	postProcesDartModelComponents(
		objectName,
		&importList,
		&attributeList,
		&constructorList,
		&fromJsonList,
		&preFromJsonList,
		&toJsonList,
		&preToJsonList,
	)
}

func postProcesDartModelComponents(
	objectName string,
	importList *[]string,
	attributeList *[]string,
	constructorList *[]string,
	fromJsonList *[]string,
	preFromJsonList *[]string,
	toJsonList *[]string,
	preToJsonList *[]string,
) {

	template := "%s\n" +
		"\n" +
		"class %s{\n" +
		"	%s\n" +
		"\n" +
		"	%s({\n" +
		"		%s\n" +
		"	});\n" +
		"\n" +
		"	Map<String,dynamic> toJson(){\n" +
		"		%s\n\n" +
		"		return {\n" +
		"			%s\n" +
		"		};\n" +
		"	}\n" +
		"\n" +
		"	factory %s.fromJson(Map<String, dynamic> data){\n" +
		"		%s\n\n" +
		"		%s %s = %s(\n" +
		"			%s\n" +
		"		);\n" +
		"		return %s;\n" +
		"	}\n" +
		"}\n"

	modelName := strings.Title(objectName) + "Model"

	finalString := formatString(template,
		strings.Join(*importList, "\n"),
		modelName,
		strings.Join(*attributeList, "\n\t"),
		modelName,
		strings.Join(*constructorList, "\n\t\t"),
		strings.Join(*preToJsonList, "\n\t\t"),
		strings.Join(*toJsonList, "\n\t\t\t"),
		modelName,
		strings.Join(*preFromJsonList, "\n\t\t"),
		modelName,
		toCamelCase(modelName),
		modelName,
		strings.Join(*fromJsonList, "\n\t\t\t"),
		toCamelCase(modelName),
	)

	writeDartFile(finalString, objectName)
}
