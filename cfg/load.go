package cfg

import "os"

type ConfigItem struct {
	Name         string   `json:"name"`
	Value        string   `json:"value"`
	Description  string   `json:"description,omitempty"`
	Type         string   `json:"type,omitempty"`
	Default      string   `json:"default,omitempty"`
	MaxValue     int      `json:"maxValue,omitempty"`
	MinValue     int      `json:"minValue,omitempty"`
	MaxLength    int      `json:"maxLength,omitempty"`
	MinLength    int      `json:"minLength,omitempty"`
	Enumerations []string `json:"enumerations,omitempty"`
}

func Load(filenames []string) ([]ConfigItem, error) {
	var allProps = map[string]string{}

	for _, filename := range filenames {
		props, err := ReadProperties(filename)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}

		for k, v := range props {
			allProps[k] = v
		}
	}

	var results []ConfigItem
	for key, value := range allProps {
		item := ConfigItem{
			Name:  key,
			Value: value,
			Type:  "string",
		}

		switch item.Name {
		case "daemon.urlpath":
			item.Description = "访问系统时的 HTTP 前缀，请确保 daemon.urlpath 值前后不要有斜杠。"
			item.MaxLength = 50
			item.MinLength = 1
		case "daemon.port":
			item.Description = "访问系统时的 HTTP 端口"
			item.Type = "integer"
			item.MinValue = 80
			item.MaxValue = 65535
		case "users.ldap_enabled":
			item.Description = "是否启用对 ldap 的支持"
			item.Type = "boolean"
		case "users.login_conflict":
			item.Enumerations = []string{"auto", "force", "disableForce"}
		}
		results = append(results, item)
	}
	return results, nil
}
