/**
 * Dynamic Swagger Category Updater
 * Updates Swagger UI category dropdowns based on database categories
 */

class DynamicSwaggerUpdater {
    constructor() {
        this.categoryEndpoint = '/api/categories/names';
        this.categories = [];
        this.initialized = false;
    }

    /**
     * Initialize the dynamic updater
     */
    async init() {
        try {
            await this.fetchCategories();
            this.updateSwaggerSpec();
            this.initialized = true;
            console.log('✅ Dynamic Swagger updater initialized with categories:', this.categories);
        } catch (error) {
            console.error('❌ Failed to initialize dynamic Swagger updater:', error);
            // Fallback to default categories
            this.categories = ['anime', 'korean-drama', 'all'];
            this.updateSwaggerSpec();
        }
    }

    /**
     * Fetch available categories from API
     */
    async fetchCategories() {
        try {
            const response = await fetch(this.categoryEndpoint);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            if (data.status === 'success' && Array.isArray(data.data)) {
                this.categories = data.data;
            } else {
                throw new Error('Invalid response format');
            }
        } catch (error) {
            console.error('Failed to fetch categories:', error);
            throw error;
        }
    }

    /**
     * Update Swagger specification with dynamic categories
     */
    updateSwaggerSpec() {
        // Check if SwaggerUIBundle is available
        if (typeof window.SwaggerUIBundle === 'undefined') {
            console.warn('SwaggerUIBundle not found, retrying in 1 second...');
            setTimeout(() => this.updateSwaggerSpec(), 1000);
            return;
        }

        // Update the swagger spec with dynamic categories
        this.injectDynamicCategories();
    }

    /**
     * Inject dynamic categories into Swagger spec
     */
    injectDynamicCategories() {
        // Wait for Swagger UI to be fully loaded
        const checkSwaggerLoaded = () => {
            const swaggerContainer = document.querySelector('.swagger-ui');
            if (!swaggerContainer) {
                setTimeout(checkSwaggerLoaded, 500);
                return;
            }

            // Update category enums in the spec
            this.updateCategoryEnums();
            
            // Monitor for new renders and update them
            this.observeSwaggerChanges();
        };

        checkSwaggerLoaded();
    }

    /**
     * Update category enum values in Swagger spec
     */
    updateCategoryEnums() {
        try {
            // Get the current Swagger spec
            const swaggerUI = window.ui;
            if (!swaggerUI) {
                console.warn('Swagger UI instance not found');
                return;
            }

            // Update the spec with dynamic categories
            const spec = swaggerUI.getState().spec.json;
            if (spec && spec.paths) {
                this.updatePathParameters(spec.paths);
            }
        } catch (error) {
            console.error('Error updating category enums:', error);
        }
    }

    /**
     * Update path parameters with dynamic categories
     */
    updatePathParameters(paths) {
        const endpointsWithCategory = [
            '/api/v1/search',
            '/api/v1/anime-terbaru',
            '/api/v1/movie',
            '/api/v1/home'
        ];

        endpointsWithCategory.forEach(endpoint => {
            if (paths[endpoint] && paths[endpoint].get && paths[endpoint].get.parameters) {
                const categoryParam = paths[endpoint].get.parameters.find(p => p.name === 'category');
                if (categoryParam && categoryParam.schema) {
                    categoryParam.schema.enum = this.categories;
                    categoryParam.schema.default = this.categories[0] || 'anime';
                }
            }
        });
    }

    /**
     * Observe Swagger UI changes and update dropdowns
     */
    observeSwaggerChanges() {
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'childList') {
                    this.updateCategoryDropdowns();
                }
            });
        });

        const swaggerContainer = document.querySelector('.swagger-ui');
        if (swaggerContainer) {
            observer.observe(swaggerContainer, {
                childList: true,
                subtree: true
            });
        }
    }

    /**
     * Update category dropdown options in the UI
     */
    updateCategoryDropdowns() {
        // Find all category parameter selects
        const categorySelects = document.querySelectorAll('select[data-param-name="category"]');
        
        categorySelects.forEach(select => {
            if (select.options.length <= 3) { // Only update if not already updated
                // Clear existing options
                select.innerHTML = '';
                
                // Add dynamic options
                this.categories.forEach(category => {
                    const option = document.createElement('option');
                    option.value = category;
                    option.textContent = this.formatCategoryName(category);
                    select.appendChild(option);
                });

                // Set default value
                if (this.categories.length > 0) {
                    select.value = this.categories[0];
                }
            }
        });
    }

    /**
     * Format category name for display
     */
    formatCategoryName(category) {
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

    /**
     * Refresh categories from server
     */
    async refresh() {
        try {
            await this.fetchCategories();
            this.updateSwaggerSpec();
            console.log('✅ Categories refreshed:', this.categories);
        } catch (error) {
            console.error('❌ Failed to refresh categories:', error);
        }
    }

    /**
     * Get current categories
     */
    getCategories() {
        return this.categories;
    }
}

// Global instance
window.dynamicSwaggerUpdater = new DynamicSwaggerUpdater();

// Auto-initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        window.dynamicSwaggerUpdater.init();
    });
} else {
    window.dynamicSwaggerUpdater.init();
}

// Expose refresh function globally for manual updates
window.refreshSwaggerCategories = () => {
    window.dynamicSwaggerUpdater.refresh();
};

// Auto-refresh every 5 minutes to pick up new categories
setInterval(() => {
    window.dynamicSwaggerUpdater.refresh();
}, 5 * 60 * 1000);

console.log('🚀 Dynamic Swagger Category Updater loaded');