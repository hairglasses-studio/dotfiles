package repo_analysis

// ImportRule maps an import path pattern to a tag and category.
type ImportRule struct {
	// Pattern is a prefix to match against Go import paths.
	Pattern  string
	Tag      string
	Category string // "framework", "protocol", "datastore", "cloud", "ai", "observability"
}

// importRules defines the mapping from Go import paths to technology tags.
// Patterns are matched as prefixes against require directives in go.mod.
var importRules = []ImportRule{
	// Frameworks
	{Pattern: "github.com/hairglasses-studio/mcpkit", Tag: "mcpkit", Category: "framework"},
	{Pattern: "github.com/mark3labs/mcp-go", Tag: "mcp-go", Category: "framework"},
	{Pattern: "github.com/modelcontextprotocol/go-sdk", Tag: "go-sdk", Category: "framework"},
	{Pattern: "github.com/charmbracelet/bubbletea", Tag: "bubbletea", Category: "framework"},
	{Pattern: "github.com/charmbracelet/lipgloss", Tag: "lipgloss", Category: "framework"},
	{Pattern: "github.com/gin-gonic/gin", Tag: "gin", Category: "framework"},
	{Pattern: "github.com/labstack/echo", Tag: "echo", Category: "framework"},
	{Pattern: "github.com/gofiber/fiber", Tag: "fiber", Category: "framework"},

	// Protocols
	{Pattern: "github.com/mark3labs/mcp-go", Tag: "mcp", Category: "protocol"},
	{Pattern: "google.golang.org/grpc", Tag: "grpc", Category: "protocol"},
	{Pattern: "github.com/a2aproject/a2a-go", Tag: "a2a", Category: "protocol"},
	{Pattern: "github.com/gorilla/websocket", Tag: "websocket", Category: "protocol"},
	{Pattern: "github.com/nats-io/nats.go", Tag: "nats", Category: "protocol"},
	{Pattern: "github.com/segmentio/kafka-go", Tag: "kafka", Category: "protocol"},
	{Pattern: "github.com/eclipse/paho.mqtt.golang", Tag: "mqtt", Category: "protocol"},

	// Datastores
	{Pattern: "github.com/mattn/go-sqlite3", Tag: "sqlite", Category: "datastore"},
	{Pattern: "modernc.org/sqlite", Tag: "sqlite", Category: "datastore"},
	{Pattern: "github.com/lib/pq", Tag: "postgres", Category: "datastore"},
	{Pattern: "github.com/jackc/pgx", Tag: "postgres", Category: "datastore"},
	{Pattern: "github.com/go-sql-driver/mysql", Tag: "mysql", Category: "datastore"},
	{Pattern: "go.mongodb.org/mongo-driver", Tag: "mongodb", Category: "datastore"},
	{Pattern: "github.com/redis/go-redis", Tag: "redis", Category: "datastore"},
	{Pattern: "go.etcd.io/bbolt", Tag: "bbolt", Category: "datastore"},

	// Cloud
	{Pattern: "github.com/aws/aws-sdk-go", Tag: "aws", Category: "cloud"},
	{Pattern: "cloud.google.com/go", Tag: "gcp", Category: "cloud"},
	{Pattern: "github.com/Azure/azure-sdk-for-go", Tag: "azure", Category: "cloud"},

	// AI
	{Pattern: "github.com/anthropics/anthropic-sdk-go", Tag: "anthropic", Category: "ai"},
	{Pattern: "github.com/openai/openai-go", Tag: "openai", Category: "ai"},
	{Pattern: "github.com/google/generative-ai-go", Tag: "gemini", Category: "ai"},

	// Observability
	{Pattern: "go.opentelemetry.io/otel", Tag: "opentelemetry", Category: "observability"},
	{Pattern: "github.com/prometheus/client_golang", Tag: "prometheus", Category: "observability"},
	{Pattern: "go.uber.org/zap", Tag: "zap", Category: "observability"},
	{Pattern: "github.com/sirupsen/logrus", Tag: "logrus", Category: "observability"},
}
