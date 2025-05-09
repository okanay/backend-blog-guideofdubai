package types

// TranslateRequest, çeviri isteği için gerekli verileri içerir
type TranslateRequest struct {
	HTML           string `json:"html"`                              // HTML içerik
	TiptapJSON     string `json:"tiptapJSON"`                        // Tiptap editör JSON içeriği
	SourceLanguage string `json:"sourceLanguage" binding:"required"` // Kaynak dil
	TargetLanguage string `json:"targetLanguage" binding:"required"` // Hedef dil
	BlogID         string `json:"blogId"`                            // İsteğe bağlı, belirli bir blogu çevirmek için
}

// TranslateResponse, çeviri yanıtı için yapı
type TranslateResponse struct {
	TranslatedHTML string `json:"translatedHTML"` // Çevrilen HTML içerik
	TranslatedJSON string `json:"translatedJSON"` // Çevrilen Tiptap JSON içerik
	SourceLanguage string `json:"sourceLanguage"` // Kaynak dil
	TargetLanguage string `json:"targetLanguage"` // Hedef dil
	TokensUsed     int    `json:"tokensUsed"`     // Kullanılan token sayısı
	Cost           any    `json:"cost"`           // Maliyet bilgisi
}

// TranslateChunkResult, bir çeviri parçasının sonucunu temsil eder
type TranslateChunkResult struct {
	Content    string
	TokensUsed int
	Error      error
}
