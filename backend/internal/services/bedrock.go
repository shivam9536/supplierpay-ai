package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// ── Bedrock LLM Client ──────────────────────

type BedrockClient struct {
	cfg    *config.Config
	logger *zap.Logger
	rt     *bedrockruntime.Client
}

func NewBedrockClient(cfg *config.Config, logger *zap.Logger) *BedrockClient {
	rt := buildBedrockRuntime(cfg, logger)
	return &BedrockClient{cfg: cfg, logger: logger, rt: rt}
}

// buildBedrockRuntime creates a bedrockruntime client.
// Priority: explicit IAM key/secret from env → default credential chain.
func buildBedrockRuntime(cfg *config.Config, logger *zap.Logger) *bedrockruntime.Client {
	region := cfg.AWSRegion
	if region == "" {
		region = "us-east-1"
	}

	// Explicit IAM key + secret from environment variables
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		logger.Info("Bedrock: using IAM key/secret from environment")
		staticCreds := credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
			awsconfig.WithCredentialsProvider(staticCreds),
		)
		if err != nil {
			logger.Error("Bedrock: failed to build AWS config with IAM keys", zap.Error(err))
		} else {
			return bedrockruntime.NewFromConfig(awsCfg)
		}
	}

	// Fallback: default credential chain (instance role, ~/.aws, etc.)
	logger.Info("Bedrock: using default AWS credential chain")
	awsCfg, _ := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
	)
	return bedrockruntime.NewFromConfig(awsCfg)
}

// ── Claude message types ─────────────────────

type claudeMessage struct {
	Role    string         `json:"role"`
	Content []claudeBlock  `json:"content"`
}

type claudeBlock struct {
	Type   string `json:"type"`
	Text   string `json:"text,omitempty"`
	Source *claudeImageSource `json:"source,omitempty"`
}

type claudeImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type claudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	Messages         []claudeMessage `json:"messages"`
	System           string          `json:"system,omitempty"`
}

type claudeResponse struct {
	Content []claudeResponseBlock `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type claudeResponseBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	// tool_use fields
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// ── Tool-use (function calling) types ────────

type claudeTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type claudeRequestWithTools struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	System           string          `json:"system,omitempty"`
	Messages         []claudeMessage `json:"messages"`
	Tools            []claudeTool    `json:"tools,omitempty"`
}

// claudeToolResultBlock is used in the tool_result content block sent back to Claude.
type claudeToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

// claudeToolUseBlock is used in the assistant message content when Claude calls a tool.
type claudeToolUseBlock struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// invokeModel calls Bedrock InvokeModel and returns the text response.
func (b *BedrockClient) invokeModel(ctx context.Context, system, userText string) (string, error) {
	req := claudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        b.cfg.BedrockMaxTokens,
		System:           system,
		Messages: []claudeMessage{
			{Role: "user", Content: []claudeBlock{{Type: "text", Text: userText}}},
		},
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	out, err := b.rt.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(b.cfg.BedrockModelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		return "", fmt.Errorf("bedrock invoke: %w", err)
	}

	var resp claudeResponse
	if err := json.Unmarshal(out.Body, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response from Bedrock")
	}
	b.logger.Info("Bedrock invoked",
		zap.String("model", b.cfg.BedrockModelID),
		zap.Int("input_tokens", resp.Usage.InputTokens),
		zap.Int("output_tokens", resp.Usage.OutputTokens),
	)
	return resp.Content[0].Text, nil
}

// invokeModelWithImage calls Bedrock with an image payload (for PDF/image invoices).
func (b *BedrockClient) invokeModelWithImage(ctx context.Context, system, userText string, imageData []byte, mimeType string) (string, error) {
	mediaType := mimeType
	if mediaType == "application/pdf" {
		// If the bytes are valid UTF-8 text (e.g. a text file uploaded as PDF),
		// send as plain text so Claude can read it directly.
		if isReadableText(imageData) {
			return b.invokeModel(ctx, system, userText+"\n\nInvoice text:\n"+string(imageData))
		}
		// For real PDFs use the Bedrock document block format.
		req := claudeRequest{
			AnthropicVersion: "bedrock-2023-05-31",
			MaxTokens:        b.cfg.BedrockMaxTokens,
			System:           system,
			Messages: []claudeMessage{
				{
					Role: "user",
					Content: []claudeBlock{
						{
							Type: "document",
							Source: &claudeImageSource{
								Type:      "base64",
								MediaType: "application/pdf",
								Data:      base64.StdEncoding.EncodeToString(imageData),
							},
						},
						{Type: "text", Text: userText},
					},
				},
			},
		}
		body, err := json.Marshal(req)
		if err != nil {
			return "", fmt.Errorf("marshal request: %w", err)
		}
		out, err := b.rt.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(b.cfg.BedrockModelID),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
			Body:        body,
		})
		if err != nil {
			// Fallback: send as plain text if document block not supported
			b.logger.Warn("PDF document block failed, falling back to text", zap.Error(err))
			return b.invokeModel(ctx, system, userText+"\n\nInvoice content (extracted):\n"+string(imageData))
		}
		var resp claudeResponse
		if err := json.Unmarshal(out.Body, &resp); err != nil || len(resp.Content) == 0 {
			return "", fmt.Errorf("empty/invalid response from Bedrock")
		}
		return resp.Content[0].Text, nil
	}

	req := claudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        b.cfg.BedrockMaxTokens,
		System:           system,
		Messages: []claudeMessage{
			{
				Role: "user",
				Content: []claudeBlock{
					{
						Type: "image",
						Source: &claudeImageSource{
							Type:      "base64",
							MediaType: mediaType,
							Data:      base64.StdEncoding.EncodeToString(imageData),
						},
					},
					{Type: "text", Text: userText},
				},
			},
		},
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	out, err := b.rt.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(b.cfg.BedrockModelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		return "", fmt.Errorf("bedrock invoke: %w", err)
	}

	var resp claudeResponse
	if err := json.Unmarshal(out.Body, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response from Bedrock")
	}
	return resp.Content[0].Text, nil
}

// ── LLMService implementation ────────────────

func (b *BedrockClient) ExtractInvoiceFields(ctx context.Context, invoiceData []byte, mimeType string) (map[string]interface{}, error) {
	b.logger.Info("Bedrock: extracting invoice fields",
		zap.String("model", b.cfg.BedrockModelID),
		zap.Int("data_size", len(invoiceData)),
	)

	system := `You are an invoice processing AI. Extract structured data from invoices and return ONLY valid JSON with no markdown fences or extra text.`

	userPrompt := `Extract the following fields from this invoice and return ONLY a JSON object:
{
  "vendor_name": "string",
  "invoice_number": "string",
  "po_reference": "string",
  "total_amount": number,
  "tax_amount": number,
  "currency": "string (e.g. INR, USD)",
  "invoice_date": "YYYY-MM-DD",
  "due_date": "YYYY-MM-DD",
  "confidence": number (0-1),
  "line_items": [
    {
      "description": "string",
      "quantity": number,
      "unit_price": number,
      "total": number
    }
  ]
}

Rules:
- Return ONLY the JSON object, no markdown, no explanation.
- Use null for missing fields.
- Dates must be YYYY-MM-DD format.
- Amounts must be numbers (not strings).`

	var text string
	var err error
	if len(invoiceData) > 0 && (strings.HasPrefix(mimeType, "image/") || mimeType == "application/pdf") {
		text, err = b.invokeModelWithImage(ctx, system, userPrompt, invoiceData, mimeType)
	} else {
		text, err = b.invokeModel(ctx, system, userPrompt+"\n\nInvoice text:\n"+string(invoiceData))
	}
	if err != nil {
		return nil, fmt.Errorf("bedrock extraction: %w", err)
	}

	// Strip markdown fences if present
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		lines := strings.Split(text, "\n")
		if len(lines) > 2 {
			text = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	var fields map[string]interface{}
	if err := json.Unmarshal([]byte(text), &fields); err != nil {
		return nil, fmt.Errorf("parse extraction response: %w (raw: %s)", err, text)
	}
	return fields, nil
}

func (b *BedrockClient) GenerateQueryEmail(ctx context.Context, invoiceDetails map[string]interface{}, discrepancies []string) (string, error) {
	b.logger.Info("Bedrock: generating query email")

	detailsJSON, _ := json.MarshalIndent(invoiceDetails, "", "  ")

	system := `You are an accounts payable specialist. Draft professional, concise supplier query emails.`

	userPrompt := fmt.Sprintf(`Draft a professional email to a supplier about invoice discrepancies.

Invoice Details:
%s

Discrepancies Found:
%s

Requirements:
- Professional and polite tone
- Clearly list each discrepancy
- Request corrected invoice or clarification
- Include subject line starting with "Subject: "
- Sign off as "SupplierPay AI — Accounts Payable"`, string(detailsJSON), strings.Join(discrepancies, "\n- "))

	return b.invokeModel(ctx, system, userPrompt)
}

func (b *BedrockClient) ExplainDecision(ctx context.Context, matchResult map[string]interface{}) (string, error) {
	b.logger.Info("Bedrock: explaining decision")

	resultJSON, _ := json.MarshalIndent(matchResult, "", "  ")

	system := `You are an AI assistant explaining invoice processing decisions to finance teams. Be concise and clear.`

	userPrompt := fmt.Sprintf(`Explain this invoice validation result in 2-3 sentences for a finance team member:

%s

Focus on: what was checked, what passed/failed, and what action was taken.`, string(resultJSON))

	return b.invokeModel(ctx, system, userPrompt)
}

// ValidateWithLLM uses the LLM to perform a semantic validation pass on top of
// the rule-based checks. It returns a structured assessment with any additional
// discrepancies the LLM identifies.
func (b *BedrockClient) ValidateWithLLM(ctx context.Context, invoiceFields map[string]interface{}, poDetails map[string]interface{}, ruleCheckSummary string) (*LLMValidationResult, error) {
	b.logger.Info("Bedrock: running LLM validation")

	invoiceJSON, _ := json.MarshalIndent(invoiceFields, "", "  ")
	poJSON, _ := json.MarshalIndent(poDetails, "", "  ")

	today := time.Now().Format("2006-01-02")

	system := `You are an expert accounts payable auditor. Analyze invoices against purchase orders and identify discrepancies. Return ONLY valid JSON.`

	userPrompt := fmt.Sprintf(`Perform a semantic validation of this invoice against the purchase order.

Today's date: %s

Invoice Fields:
%s

Purchase Order Details:
%s

Rule-Based Check Summary (already performed):
%s

IMPORTANT RULES — do NOT flag the following as issues:
- Invoice dates or due dates that are in the past (historical invoices are normal)
- PO numbers whose year component differs from the invoice year (PO numbering conventions vary)
- Round-number amounts on service contracts (common for fixed-price engagements)
- GST/tax at standard Indian rates (18%%, 12%%, 5%%) applied on top of PO subtotal values
- Items already validated as PASSED in the Rule-Based Check Summary above

Analyze ONLY for:
1. Semantic mismatches in line-item descriptions vs PO scope (e.g. "Cloud Hosting" billed against a "Security Audit" PO)
2. Invoice total genuinely exceeding the PO remaining value (hard math check)
3. Duplicate invoice numbers (already caught by rules — only flag if rules missed it)
4. Missing mandatory fields (vendor name, invoice number, PO reference, total amount)
5. Genuine fraud signals (e.g. altered amounts, mismatched vendor details)

Return ONLY this JSON:
{
  "overall_assessment": "APPROVE" | "FLAG" | "REJECT",
  "confidence": number (0-1),
  "additional_discrepancies": ["string", ...],
  "semantic_matches": [{"invoice_desc": "string", "po_desc": "string", "match": true|false, "note": "string"}],
  "risk_flags": ["string", ...],
  "explanation": "string (2-3 sentences)"
}`, today, string(invoiceJSON), string(poJSON), ruleCheckSummary)

	text, err := b.invokeModel(ctx, system, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("bedrock LLM validation: %w", err)
	}

	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		lines := strings.Split(text, "\n")
		if len(lines) > 2 {
			text = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	var result LLMValidationResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse LLM validation response: %w (raw: %s)", err, text)
	}
	return &result, nil
}

// LLMValidationResult holds the structured output from the LLM validation pass.
type LLMValidationResult struct {
	OverallAssessment       string                   `json:"overall_assessment"`
	Confidence              float64                  `json:"confidence"`
	AdditionalDiscrepancies []string                 `json:"additional_discrepancies"`
	SemanticMatches         []SemanticMatch          `json:"semantic_matches"`
	RiskFlags               []string                 `json:"risk_flags"`
	Explanation             string                   `json:"explanation"`
}

type SemanticMatch struct {
	InvoiceDesc string `json:"invoice_desc"`
	PODesc      string `json:"po_desc"`
	Match       bool   `json:"match"`
	Note        string `json:"note"`
}

// isReadableText returns true if the byte slice is valid UTF-8 text
// (i.e. a plain-text file uploaded with a PDF mime type).
func isReadableText(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	// Real PDFs always start with %PDF
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return false
	}
	// Check if it's valid UTF-8 with mostly printable characters
	text := string(data)
	printable := 0
	for _, r := range text {
		if r == '\n' || r == '\r' || r == '\t' || (r >= 32 && r < 127) {
			printable++
		}
	}
	return len(text) > 0 && float64(printable)/float64(len(text)) > 0.85
}

// ── Mock LLM Client (for local dev) ─────────

type MockLLMClient struct {
	logger *zap.Logger
}

func NewMockLLMClient(logger *zap.Logger) *MockLLMClient {
	return &MockLLMClient{logger: logger}
}

func (m *MockLLMClient) ExtractInvoiceFields(ctx context.Context, invoiceData []byte, mimeType string) (map[string]interface{}, error) {
	m.logger.Info("[MOCK] Extracting invoice fields")

	fields := map[string]interface{}{
		"vendor_name":    "Acme Corp",
		"invoice_number": "INV-2026-001",
		"po_reference":   "PO-2026-100",
		"total_amount":   50000.00,
		"tax_amount":     9000.00,
		"currency":       "INR",
		"invoice_date":   "2026-03-01",
		"due_date":       "2026-03-31",
		"confidence":     0.95,
		"line_items": []map[string]interface{}{
			{"description": "Cloud Hosting - March", "quantity": 1, "unit_price": 41000.00, "total": 41000.00},
			{"description": "Support Services", "quantity": 1, "unit_price": 9000.00, "total": 9000.00},
		},
	}
	return fields, nil
}

func (m *MockLLMClient) GenerateQueryEmail(ctx context.Context, invoiceDetails map[string]interface{}, discrepancies []string) (string, error) {
	m.logger.Info("[MOCK] Generating query email")

	detailsJSON, _ := json.MarshalIndent(invoiceDetails, "", "  ")

	email := fmt.Sprintf(`Subject: Invoice Discrepancy — Action Required

Dear Supplier,

We have identified the following discrepancies in your recent invoice:

%s

Invoice Details:
%s

Please review and provide a corrected invoice or clarification at your earliest convenience.

Best regards,
SupplierPay AI — Accounts Payable`, discrepancies, string(detailsJSON))

	return email, nil
}

func (m *MockLLMClient) ExplainDecision(ctx context.Context, matchResult map[string]interface{}) (string, error) {
	m.logger.Info("[MOCK] Explaining decision")
	return "Invoice auto-approved: PO matched, amounts verified, line items consistent.", nil
}

func (m *MockLLMClient) ValidateWithLLM(ctx context.Context, invoiceFields map[string]interface{}, poDetails map[string]interface{}, ruleCheckSummary string) (*LLMValidationResult, error) {
	m.logger.Info("[MOCK] LLM validation")
	return &LLMValidationResult{
		OverallAssessment:       "PASS",
		Confidence:              0.95,
		AdditionalDiscrepancies: []string{},
		SemanticMatches:         []SemanticMatch{},
		RiskFlags:               []string{},
		Explanation:             "Mock LLM validation: all checks passed, invoice appears legitimate.",
	}, nil
}

