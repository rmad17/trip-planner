package documents

import "github.com/gin-gonic/gin"

// RouterGroupDocuments sets up routes for documents nested under trip plans
func RouterGroupDocuments(router *gin.RouterGroup) {
	// Documents nested under Trip Plans
	router.GET("/:id/documents", GetDocuments)    // GET /trip/:id/documents
	router.POST("/:id/documents", UploadDocument) // POST /trip/:id/documents
}

// RouterGroupDocumentItems sets up CRUD routes for individual documents
func RouterGroupDocumentItems(router *gin.RouterGroup) {
	router.GET("/:id", GetDocument)               // GET /documents/:id
	router.PUT("/:id", UpdateDocument)            // PUT /documents/:id
	router.DELETE("/:id", DeleteDocument)         // DELETE /documents/:id
	router.GET("/:id/download", DownloadDocument) // GET /documents/:id/download
}
