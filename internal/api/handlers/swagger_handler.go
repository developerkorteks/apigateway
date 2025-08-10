package handlers

import (
	"apicategorywithfallback/internal/service"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SwaggerHandler handles custom Swagger UI with dynamic categories
type SwaggerHandler struct {
	apiService *service.APIService
}

// NewSwaggerHandler creates a new swagger handler
func NewSwaggerHandler(apiService *service.APIService) *SwaggerHandler {
	return &SwaggerHandler{
		apiService: apiService,
	}
}

// ServeSwaggerUI serves custom Swagger UI with dynamic category injection
func (h *SwaggerHandler) ServeSwaggerUI(c *gin.Context) {
	// Get available categories
	categories, err := h.apiService.GetCategoryNames()
	if err != nil {
		// Fallback to default categories
		categories = []string{"anime", "korean-drama", "all"}
	}

	// Custom Swagger HTML template with dynamic category injection
	swaggerHTML := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>API Fallback Service - Dynamic Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .dynamic-category-info {
            background: #e8f5e8;
            border: 1px solid #4caf50;
            border-radius: 4px;
            padding: 10px;
            margin: 10px 0;
            font-family: monospace;
        }
        .category-list {
            display: flex;
            flex-wrap: wrap;
            gap: 5px;
            margin-top: 5px;
        }
        .category-tag {
            background: #4caf50;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    
    <!-- Dynamic Category Info -->
    <div class="dynamic-category-info">
        <strong>🎯 Dynamic Categories Loaded:</strong>
        <div class="category-list">
            {{range .Categories}}
            <span class="category-tag">{{.}}</span>
            {{end}}
        </div>
        <small>Categories are automatically updated when you add new ones via dashboard!</small>
    </div>

    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        // Dynamic categories from server
        window.DYNAMIC_CATEGORIES = {{.CategoriesJSON}};
        
        // Initialize Swagger UI
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/doc.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                onComplete: function() {
                    // Inject dynamic categories after Swagger loads
                    injectDynamicCategories();
                }
            });
            
            window.ui = ui;
        };

        // Function to inject dynamic categories
        function injectDynamicCategories() {
            console.log('🎯 Injecting dynamic categories:', window.DYNAMIC_CATEGORIES);
            
            // Update category parameter dropdowns
            setTimeout(() => {
                updateCategoryDropdowns();
            }, 1000);
            
            // Set up observer for dynamic updates
            observeSwaggerChanges();
        }

        // Update category dropdown options
        function updateCategoryDropdowns() {
            const categorySelects = document.querySelectorAll('select[data-param-name="category"]');
            
            categorySelects.forEach(select => {
                // Clear existing options
                select.innerHTML = '';
                
                // Add dynamic options
                window.DYNAMIC_CATEGORIES.forEach(category => {
                    const option = document.createElement('option');
                    option.value = category;
                    option.textContent = formatCategoryName(category);
                    select.appendChild(option);
                });

                // Set default value
                if (window.DYNAMIC_CATEGORIES.length > 0) {
                    select.value = window.DYNAMIC_CATEGORIES[0];
                }
            });
        }

        // Format category name for display
        function formatCategoryName(category) {
            const formatMap = {
                'anime': 'Anime',
                'korean-drama': 'Korean Drama',
                'donghua': 'Donghua (Chinese Animation)',
                'film': 'Film/Movie',
                'manhwa': 'Manhwa/Webtoon',
                'all': 'All Categories'
            };

            return formatMap[category] || category.charAt(0).toUpperCase() + category.slice(1);
        }

        // Observe Swagger UI changes
        function observeSwaggerChanges() {
            const observer = new MutationObserver((mutations) => {
                mutations.forEach((mutation) => {
                    if (mutation.type === 'childList') {
                        updateCategoryDropdowns();
                    }
                });
            });

            const swaggerContainer = document.querySelector('#swagger-ui');
            if (swaggerContainer) {
                observer.observe(swaggerContainer, {
                    childList: true,
                    subtree: true
                });
            }
        }

        // Auto-refresh categories every 5 minutes
        setInterval(async () => {
            try {
                const response = await fetch('/api/categories/names');
                const data = await response.json();
                if (data.status === 'success') {
                    window.DYNAMIC_CATEGORIES = data.data;
                    updateCategoryDropdowns();
                    console.log('✅ Categories refreshed:', window.DYNAMIC_CATEGORIES);
                }
            } catch (error) {
                console.error('❌ Failed to refresh categories:', error);
            }
        }, 5 * 60 * 1000);

        console.log('🚀 Dynamic Swagger UI initialized with categories:', window.DYNAMIC_CATEGORIES);
    </script>
</body>
</html>
`

	// Parse template
	tmpl, err := template.New("swagger").Parse(swaggerHTML)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template parsing error: %v", err)
		return
	}

	// Convert categories to JSON for JavaScript
	categoriesJSON := `[`
	for i, cat := range categories {
		if i > 0 {
			categoriesJSON += `,`
		}
		categoriesJSON += `"` + cat + `"`
	}
	categoriesJSON += `]`

	// Template data
	data := struct {
		Categories     []string
		CategoriesJSON template.JS
	}{
		Categories:     categories,
		CategoriesJSON: template.JS(categoriesJSON),
	}

	// Set content type
	c.Header("Content-Type", "text/html; charset=utf-8")

	// Execute template
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Template execution error: %v", err)
		return
	}
}
