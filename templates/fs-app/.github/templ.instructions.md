---
applyTo: "**/*.templ"
---

# GitHub Copilot Instructions for .templ Files

You are working with templ, a Go-based HTML rendering library for building server-side rendered web applications. These instructions will help you generate high-quality, secure, and idiomatic templ code.

## File Structure and Syntax

### Basic Template Structure

- Start with a Go package declaration and imports
- Define components using the `templ` keyword followed by the component name and parameters
- Components are compiled into Go functions that return `templ.Component`
- Use curly braces `{}` for Go expressions and variables

```templ
package main

import "fmt"

templ headerTemplate(name string) {
    <header data-testid="headerTemplate">
        <h1>{ name }</h1>
        <p>Welcome to { fmt.Sprintf("Hello %s", name) }</p>
    </header>
}
```

### Component Naming and Visibility

- Follow Go visibility rules: uppercase names are exported (public), lowercase are private
- Use descriptive names that indicate the component's purpose
- Prefer camelCase for component names

### Parameter Types

- Components can accept any Go types as parameters
- Use struct types for complex parameter sets
- Consider using interfaces when appropriate for flexibility

## Security Best Practices

### Automatic Escaping

- templ automatically HTML-escapes all dynamic content to prevent XSS attacks
- Text expressions in `{}` are automatically escaped using `templ.EscapeString`
- No manual escaping needed for regular content

### Script and Style Restrictions

- `<script>` and `<style>` tags cannot contain dynamic variables or expressions
- Only constant CSS and JavaScript are allowed in these sections
- Use external files or `templ.ComponentScript` for dynamic scripts

### Event Handlers

- Use `templ.ComponentScript` for `on*` attributes (onClick, onSubmit, etc.)
- Never directly embed user data in event handlers

### URLs and Links

- Use `templ.SafeURL` for `href` attributes to prevent JavaScript injection
- URLs are automatically sanitized to remove potential attacks

### CSS Classes and Styles

- CSS class names are automatically sanitized
- Unsafe class names are replaced with `--templ-css-class-safe-name`
- Style attributes cannot be dynamic expressions - use CSS classes instead

## Component Patterns

### Simple Components

```templ
templ button(text string, disabled bool) {
    <button disabled?={ disabled } class="btn">
        { text }
    </button>
}
```

### Conditional Rendering

```templ
templ userProfile(user User, isLoggedIn bool) {
    if isLoggedIn {
        <div class="profile">
            <h2>{ user.Name }</h2>
            <p>{ user.Email }</p>
        </div>
    } else {
        <div class="login-prompt">
            <p>Please log in to view your profile</p>
        </div>
    }
}
```

### Loops and Iteration

```templ
templ itemList(items []Item) {
    <ul>
        for _, item := range items {
            <li>
                <h3>{ item.Title }</h3>
                <p>{ item.Description }</p>
            </li>
        }
    </ul>
}
```

### Component Composition

```templ
templ layout(title string, content templ.Component) {
    <!DOCTYPE html>
    <html>
        <head>
            <title>{ title }</title>
        </head>
        <body>
            @header(title)
            <main>
                @content
            </main>
            @footer()
        </body>
    </html>
}
```

## Development Workflow

### Code Generation

- Always run `templ generate` after modifying .templ files
- Use `templ generate -watch` during development for automatic regeneration
- Generated Go files should not be manually edited

### Formatting

- Use `templ fmt` to format template files consistently
- Format before committing code
- Consider using `templ fmt -fail` in CI/CD pipelines

### Testing

- Components can be tested by calling their generated functions
- Use `strings.Builder` or similar to capture rendered output for assertions
- Test both the HTML structure and dynamic content

## Integration Patterns

### HTTP Handlers

```go
func handler(w http.ResponseWriter, r *http.Request) {
    component := myTemplate("Hello World")
    component.Render(context.Background(), w)
}
```

### Framework Integration

- Compatible with Chi, Echo, Gin, Go Fiber, and other Go web frameworks
- Use framework-specific adapters when available
- Components implement `templ.Component` interface with `Render` method

### CSRF Protection

```templ
templ form() {
    <form method="post">
        <input type="hidden" name="csrf_token" value={ csrf.Token(r) } />
        <!-- form fields -->
    </form>
}
```

## Best Practices

### Performance

- Prefer server-side rendering over client-side JavaScript when possible
- Use static components for unchanging content
- Consider component caching for expensive operations

### Maintainability

- Keep components small and focused on single responsibilities
- Use composition over large monolithic components
- Group related components in the same file or package

### Accessibility

- Always include proper semantic HTML elements
- Add ARIA attributes where necessary
- Include alt text for images and proper form labels

### Error Handling

- Handle potential nil values in template parameters
- Use Go's error handling patterns in component logic
- Provide fallback content for error states

## Common Patterns to Avoid

- Don't put complex business logic in templates - use Go functions instead
- Don't create deeply nested component hierarchies
- Don't ignore the automatic escaping - it's there for security
- Don't manually concatenate HTML strings - use components
- Don't use dynamic values in `<script>` or `<style>` tags

## IDE and Tooling

- Use the templ Language Server (`templ lsp`) for syntax highlighting and completion
- Configure your editor to run `templ generate` on file save
- Use `templ fmt` as a pre-commit hook
- Integrate with gopls for Go code completion within templates

## Advanced Component Patterns

### Method Components (Attached to Types)

Components can be defined as methods on Go types for better encapsulation:

```templ
type Person struct {
    Name  string
    Email string
}

templ (p Person) bioTemplate() {
    <div class="bio">
        <h3>{ p.Name }</h3>
        <p>Email: { p.Email }</p>
    </div>
}
```

### Code-Only Components

For advanced use cases, implement the `templ.Component` interface manually:

```go
type customComponent struct {
    content string
}

func (c customComponent) Render(ctx context.Context, w io.Writer) error {
    // Manual HTML escaping required
    _, err := w.Write([]byte(templ.EscapeString(c.content)))
    return err
}
```

## Server-Side Rendering Patterns

### Static Page Pattern

```go
// For static components without dynamic data
http.Handle("/", templ.Handler(hello()))
http.ListenAndServe(":8080", nil)
```

### Dynamic Data Pattern

```go
// For components requiring request-specific data
http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    user := getUserByID(userID)
    userProfile(user).Render(r.Context(), w)
})
```

### HTTP Configuration Options

```go
// Set custom status codes and headers
templ.WithStatus(404)(notFoundComponent).Render(ctx, w)
templ.WithContentType("text/plain")(plainTextComponent).Render(ctx, w)
templ.WithErrorHandler(customErrorHandler)(component).Render(ctx, w)
```

## Static Site Generation

### File Generation Pattern

```go
func generateStaticSite() error {
    pages := []struct {
        filename string
        component templ.Component
    }{
        {"index.html", homePage("Welcome")},
        {"about.html", aboutPage()},
        {"contact.html", contactPage()},
    }

    for _, page := range pages {
        f, err := os.Create(page.filename)
        if err != nil {
            return err
        }
        defer f.Close()

        err = page.component.Render(context.Background(), f)
        if err != nil {
            return err
        }
    }
    return nil
}
```

## CLI Commands and Development Workflow

### Complete CLI Command Reference

#### Generation Commands

```bash
# Generate all .templ files
templ generate

# Generate specific file
templ generate -f header.templ

# Watch mode for development
templ generate --watch

# Lazy generation (only if source is newer)
templ generate --lazy

# Set worker count for parallel processing
templ generate --workers 4

# Generate with source map visualization
templ generate --source-map
```

#### Formatting Commands

```bash
# Format all files
templ fmt .

# Format with CI validation (exit code 1 if formatting needed)
templ fmt --fail .

# Format specific directory
templ fmt ./components
```

#### Development Tools

```bash
# Start Language Server Protocol
templ lsp

# LSP with logging
templ lsp --log /path/to/lsp.log

# Environment information
templ info

# Version information
templ version
```

## AWS Lambda Deployment

### Optimized Build Configuration

```bash
# Build for Lambda with optimizations
GOOS=linux GOARCH=arm64 go build \
  -ldflags "-s -w" \
  -tags lambda.norpc \
  -o bootstrap main.go
```

### Lambda Handler Pattern

```go
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/events"
    "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    mux := http.NewServeMux()
    mux.Handle("/", templ.Handler(homePage()))

    return httpadapter.New(mux).ProxyWithContext(ctx, req)
}
```

### Recommended Lambda Settings

- Runtime: Amazon Linux 2
- Memory: 1024 MB
- Architecture: ARM64
- Use CloudFront + S3 for static assets

## Framework Integration Examples

### Chi Router

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Get("/", func(w http.ResponseWriter, r *http.Request) {
    homePage().Render(r.Context(), w)
})
```

### Echo Framework

```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.GET("/", func(c echo.Context) error {
    return homePage().Render(c.Request().Context(), c.Response().Writer)
})
```

### Gin Framework

```go
import "github.com/gin-gonic/gin"

r := gin.Default()
r.GET("/", func(c *gin.Context) {
    homePage().Render(c.Request.Context(), c.Writer)
})
```

### Go Fiber

```go
import "github.com/gofiber/fiber/v2"

app := fiber.New()
app.Get("/", func(c *fiber.Ctx) error {
    c.Set("Content-Type", "text/html")
    return homePage().Render(c.Context(), c.Response().BodyWriter())
})
```

## Component Libraries Integration

### templUI Usage

```go
import (
    "github.com/axzilla/templui/components"
    "github.com/axzilla/templui/icons"
)

templ myPage() {
    <div class="container mx-auto p-4">
        @components.Button(components.ButtonProps{
            Text: "Submit Form",
            Variant: "primary",
            Size: "lg",
            IconRight: icons.ArrowRight(icons.IconProps{Size: "16"}),
            OnClick: submitHandler(),
        })

        @components.Card(components.CardProps{
            Title: "User Information",
            Children: userForm(),
        })
    </div>
}
```

## Testing Patterns

### Component Testing

```go
func TestHeaderTemplate(t *testing.T) {
    var buf bytes.Buffer
    component := headerTemplate("Test User")

    err := component.Render(context.Background(), &buf)
    if err != nil {
        t.Fatal(err)
    }

    html := buf.String()
    if !strings.Contains(html, "Test User") {
        t.Error("Expected 'Test User' in rendered HTML")
    }
    if !strings.Contains(html, "<header") {
        t.Error("Expected header tag in rendered HTML")
    }
}
```

### Integration Testing

```go
func TestHTTPHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()

    handler := func(w http.ResponseWriter, r *http.Request) {
        homePage("Welcome").Render(r.Context(), w)
    }

    handler(w, req)

    if w.Code != 200 {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

## Migration and Legacy Support

### Legacy Syntax Migration

For versions ≤ v0.2.663, use the migration command:

```bash
templ migrate
```

Old syntax: `{%= name %}`
New syntax: `{ name }`

## Performance Optimization

### Component Caching

```go
var componentCache = make(map[string]templ.Component)

func cachedComponent(key string, generator func() templ.Component) templ.Component {
    if cached, exists := componentCache[key]; exists {
        return cached
    }
    component := generator()
    componentCache[key] = component
    return component
}
```

### Lazy Loading

```templ
templ lazySection(shouldLoad bool, data interface{}) {
    if shouldLoad {
        @expensiveComponent(data)
    } else {
        <div class="placeholder">Loading...</div>
    }
}
```

## Error Handling Patterns

### Safe Rendering with Error Boundaries

```templ
templ safeComponent(data *Data) {
    if data == nil {
        <div class="error">No data available</div>
        return
    }

    <div class="content">
        { data.Title }
    </div>
}
```

### Fallback Content

```templ
templ userWidget(user *User) {
    if user != nil {
        <div class="user-info">
            <h3>{ user.Name }</h3>
            <p>{ user.Email }</p>
        </div>
    } else {
        <div class="user-placeholder">
            <p>Please log in to view profile</p>
        </div>
    }
}
```

## Security Implementation Details

### Content Security Policy (CSP)

```go
func cspHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
        next.ServeHTTP(w, r)
    })
}
```

### CSRF Token Integration

```templ
templ secureForm(csrfToken string) {
    <form method="POST" action="/submit">
        <input type="hidden" name="_csrf" value={ csrfToken } />
        <input type="text" name="username" required />
        <input type="password" name="password" required />
        <button type="submit">Login</button>
    </form>
}
```

Remember: templ combines the power of Go with the expressiveness of HTML templates while maintaining compile-time safety and automatic security protections. Use these patterns to build robust, secure, and maintainable web applications.
