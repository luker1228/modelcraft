package graphql

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"modelcraft/pkg/httpheader"
)

// PlaygroundConfig 配置 GraphQL Playground
type PlaygroundConfig struct {
	Endpoint string // GraphQL 端点 URL
	Title    string // 页面标题
}

// playgroundTemplate HTML 模板
const playgroundTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
    <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
    <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
    <div id="root"></div>
    <script>
        window.addEventListener('load', function (event) {
            GraphQLPlayground.init(document.getElementById('root'), {
                endpoint: '{{.Endpoint}}',
                settings: {
                    'request.credentials': 'same-origin',
                    'editor.theme': 'dark',
                    'editor.cursorShape': 'line',
                    'editor.reuseHeaders': true,
                    'tracing.hideTracingResponse': true,
                    'queryPlan.hideQueryPlanResponse': true,
                    'editor.fontSize': 14,
                    'editor.fontFamily': '"Source Code Pro", "Consolas", "Inconsolata", ' +
                        '"Droid Sans Mono", "Monaco", monospace',
                    'request.credentials': 'same-origin'
                }
            })
        })
    </script>
</body>
</html>`

// Handler returns a Gin handler for GraphQL Playground.
func Handler(config PlaygroundConfig) gin.HandlerFunc {
	// 解析模板
	tmpl, err := template.New("playground").Parse(playgroundTemplate)
	if err != nil {
		// 如果模板解析失败,返回错误处理器
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to initialize GraphQL Playground",
			})
		}
	}

	return func(c *gin.Context) {
		// 设置响应头
		c.Header(httpheader.ContentType, httpheader.ContentTypeTextHTMLUTF8)
		c.Status(http.StatusOK)

		// 渲染模板
		if err := tmpl.Execute(c.Writer, config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to render GraphQL Playground",
			})
		}
	}
}

// HTTPHandler returns a net/http handler for GraphQL Playground.
func HTTPHandler(config PlaygroundConfig) http.HandlerFunc {
	tmpl, err := template.New("playground").Parse(playgroundTemplate)
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Failed to initialize GraphQL Playground", http.StatusInternalServerError)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpheader.ContentType, httpheader.ContentTypeTextHTMLUTF8)
		w.WriteHeader(http.StatusOK)
		if err := tmpl.Execute(w, config); err != nil {
			http.Error(w, "Failed to render GraphQL Playground", http.StatusInternalServerError)
		}
	}
}
