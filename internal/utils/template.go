package utils

import (
	"bytes"
	"fmt"
	"text/template"
)

// ProcessTemplate 处理模板字符串，使用提供的变量进行替换
func ProcessTemplate(templateStr string, variables map[string]interface{}) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}

	return buf.String(), nil
}
