# Documentation System

This project includes a comprehensive documentation system using **Hugo** + **gomarkdoc** - providing MkDocs-style functionality for Go projects.

## 🚀 Quick Start

```bash
# Install documentation tools
make docs-setup

# Start development server with live reload
make docs-serve
# Visit: http://localhost:1313

# Build static site for deployment
make docs-build
```

## 📖 Documentation Structure

```
docs/
├── config.yaml              # Hugo configuration
├── content/                 # Documentation content
│   ├── _index.md           # Homepage
│   ├── docs/               # User guides and tutorials
│   │   ├── getting-started.md
│   │   ├── architecture.md
│   │   └── deployment.md
│   ├── examples/           # Code examples
│   └── api/                # Auto-generated API docs (from gomarkdoc)
├── static/                 # Static assets (images, etc.)
├── themes/                 # Hugo themes
└── public/                 # Generated site (after hugo build)
```

## 🔧 How It Works

### 1. Hugo Static Site Generator
- **Written in Go** - Perfect fit for Go projects
- **Fast builds** - Millisecond rebuild times
- **Live reload** - Instant preview during development  
- **Modern themes** - Responsive, documentation-focused designs

### 2. gomarkdoc API Generation
- **Automatic Go docs** - Generates markdown from Go code and comments
- **Multiple formats** - Supports GitHub, GitLab, and plain markdown
- **Custom templates** - Customizable output format
- **Build integration** - Runs as part of documentation build

### 3. Docsy Theme
- **Documentation-focused** - Designed specifically for technical docs
- **Search integration** - Built-in site search
- **Navigation** - Sidebar navigation with breadcrumbs
- **Mobile responsive** - Works great on all devices
- **GitHub integration** - Edit page links, version display

## 📝 Writing Documentation

### Adding New Pages

1. **Create markdown file** in `docs/content/docs/`:
   ```bash
   # Create new guide
   touch docs/content/docs/my-feature.md
   ```

2. **Add frontmatter** with metadata:
   ```yaml
   ---
   title: "My Feature Guide"
   linkTitle: "My Feature"
   weight: 20
   description: "How to use my feature"
   ---
   
   # My Feature Guide
   Content here...
   ```

3. **Preview changes**:
   ```bash
   make docs-serve
   ```

### API Documentation

API documentation is automatically generated from your Go code:

```go
// Package mypackage provides utilities for data processing.
//
// This package includes functions for parsing, validating, and
// transforming data in various formats.
package mypackage

// ProcessData processes input data and returns the result.
//
// The function accepts raw data and applies the following transformations:
//   - Validation of input format
//   - Data cleaning and normalization  
//   - Output formatting
//
// Example usage:
//   result, err := ProcessData(rawData)
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println(result)
func ProcessData(data []byte) (string, error) {
    // Implementation...
}
```

Run `make docs-generate` to update API documentation.

### Code Examples

Include runnable code examples:

````markdown
## Example Usage

```go
package main

import (
    "fmt"
    "github.com/your-org/go-template-project/internal/shared"
)

func main() {
    config := shared.LoadConfig()
    fmt.Printf("App: %s v%s\n", config.Name, config.Version)
}
```

Build and run:
```bash
go run ./cmd/cli
```
````

## 🎨 Customization

### Theme Configuration

Edit `docs/config.yaml` to customize:

```yaml
params:
  # Site colors and branding
  ui:
    navbar_logo: true
    sidebar_search_disable: false
    
  # GitHub integration  
  github_repo: 'https://github.com/your-org/go-template-project'
  edit_page: true
  
  # Version information
  version: '1.0.0'
  version_menu: true
```

### Custom Styling

Add custom CSS in `docs/static/css/`:

```css
/* docs/static/css/custom.css */
.td-navbar {
    background: linear-gradient(90deg, #1e3a8a, #3b82f6);
}

.td-sidebar-nav {
    border-right: 1px solid #e5e7eb;
}
```

Reference in config:
```yaml
params:
  custom_css: ["css/custom.css"]
```

### Navigation Menu

Configure navigation in `docs/config.yaml`:

```yaml
menu:
  main:
    - name: "Documentation"
      url: "/docs/"
      weight: 10
    - name: "API Reference"
      url: "/api/"
      weight: 20
    - name: "Examples" 
      url: "/examples/"
      weight: 30
```

## 🚀 Deployment

### GitHub Pages

1. **Enable GitHub Pages** in repository settings
2. **Add workflow** (already included in `.github/workflows/docs.yml`):
   ```yaml
   name: Build and Deploy Docs
   on:
     push:
       branches: [ main ]
   jobs:
     docs:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - name: Setup Hugo
           uses: peaceiris/actions-hugo@v2
           with:
             hugo-version: 'latest'
         - name: Build docs
           run: make docs-build
         - name: Deploy to GitHub Pages
           uses: peaceiris/actions-gh-pages@v3
           with:
             github_token: ${{ secrets.GITHUB_TOKEN }}
             publish_dir: ./docs/public
   ```

### Custom Domain

Add `docs/static/CNAME` with your domain:
```
docs.yourproject.com
```

### Docker Deployment

The documentation can be served with a simple nginx container:

```dockerfile
# docs.Dockerfile
FROM nginx:alpine
COPY docs/public /usr/share/nginx/html
EXPOSE 80
```

Build and run:
```bash
make docs-build
docker build -f docs.Dockerfile -t project-docs .
docker run -p 8080:80 project-docs
```

## 🔄 Development Workflow

### Daily Documentation

1. **Write documentation** alongside code changes
2. **Preview locally** with `make docs-serve`  
3. **Update API docs** with `make docs-generate`
4. **Commit and push** - docs auto-deploy via GitHub Actions

### Documentation Review

- **Link checking** - Hugo validates internal links
- **Spelling/grammar** - Use tools like `aspell` or `grammarly`
- **Mobile testing** - Check responsive design on different devices
- **Performance** - Hugo builds are typically <100ms

## 📚 Best Practices

### Content Organization
- **Logical hierarchy** - Use clear folder structure
- **Cross-references** - Link related documentation
- **Code examples** - Include working, tested examples
- **Screenshots** - Add visuals for UI-related topics

### API Documentation
- **Clear descriptions** - Explain what, why, and how
- **Examples** - Show realistic usage patterns
- **Error handling** - Document possible errors and solutions
- **Performance notes** - Mention any performance considerations

### Maintenance
- **Regular updates** - Keep docs current with code changes
- **Dead link checking** - Validate external links periodically
- **User feedback** - Monitor issues and questions for doc improvements
- **Analytics** - Track popular pages to guide content priorities

## 🆚 Comparison with MkDocs

| Feature | Hugo + gomarkdoc | MkDocs |
|---------|------------------|---------|
| **Language** | Go (native) | Python |
| **Speed** | Very fast builds | Moderate speed |
| **Themes** | Many options | Material theme popular |
| **Go Integration** | Excellent (gomarkdoc) | Limited |
| **Plugin Ecosystem** | Hugo modules | Rich plugin system |
| **Learning Curve** | Moderate | Easy |
| **GitHub Integration** | Native | Via plugins |

## 🛠️ Troubleshooting

### Common Issues

**Hugo not found:**
```bash
# Install Hugo
brew install hugo  # macOS
# or
sudo apt-get install hugo  # Ubuntu
# or
sudo yum install hugo  # CentOS/RHEL
```

**Theme not loading:**
```bash
# Initialize Hugo modules
cd docs
hugo mod init github.com/your-org/go-template-project/docs
hugo mod get github.com/google/docsy@latest
```

**API docs not generating:**
```bash
# Install gomarkdoc
go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest

# Check Go modules
go mod tidy
```

**Live reload not working:**
```bash
# Ensure Hugo server binds to all interfaces
cd docs && hugo server --bind 0.0.0.0
```

## 📞 Support

- **Hugo Documentation**: https://gohugo.io/documentation/
- **Docsy Theme Guide**: https://www.docsy.dev/docs/
- **gomarkdoc Usage**: https://github.com/princjef/gomarkdoc
- **Project Issues**: https://github.com/your-org/go-template-project/issues

---

**Ready to document?** Start with `make docs-setup` and then `make docs-serve`!