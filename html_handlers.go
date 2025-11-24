package main

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// generateFaviconDataUrl 生成简单的SVG favicon
func generateFaviconDataUrl() string {
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32"><circle cx="16" cy="16" r="16" fill="#f3f4f6"/><text x="16" y="22" font-family="Arial, sans-serif" font-size="20" font-weight="bold" fill="#4285f4" text-anchor="middle">Y</text></svg>`
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svgContent))
}

const indexHtml = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>y-router - Claude API 协议转换器</title>
    <link rel="shortcut icon" type="image/svg+xml" href="%s">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'PingFang SC', 'Microsoft YaHei', sans-serif;
            line-height: 1.6;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(45deg, #2c3e50, #3498db);
            color: white;
            text-align: center;
            padding: 40px 20px;
        }

        .header h1 {
            font-size: 2.2em;
            margin-bottom: 10px;
            font-weight: 300;
        }

        .header p {
            font-size: 1.1em;
            opacity: 0.9;
        }

        .content {
            padding: 40px;
        }

        .step {
            margin-bottom: 30px;
            padding: 20px;
            border-left: 4px solid #3498db;
            background: #f8f9fa;
            border-radius: 0 8px 8px 0;
        }

        .step h2 {
            color: #2c3e50;
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            font-size: 1.3em;
        }

        .step-number {
            background: #3498db;
            color: white;
            width: 28px;
            height: 28px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 15px;
            font-weight: bold;
            font-size: 0.9em;
        }

        .code-block {
            background: #2c3e50;
            color: #ecf0f1;
            padding: 15px;
            border-radius: 6px;
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            margin: 15px 0;
            overflow-x: auto;
            font-size: 0.9em;
            position: relative;
        }

        .code-block-wrapper {
            position: relative;
        }

        .copy-button {
            position: absolute;
            top: 10px;
            right: 10px;
            background: #3498db;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            opacity: 0.8;
            transition: opacity 0.2s;
        }

        .copy-button:hover {
            opacity: 1;
            background: #2980b9;
        }

        .copy-button.copied {
            background: #27ae60;
        }

        .success {
            background: linear-gradient(45deg, #27ae60, #2ecc71);
            color: white;
            padding: 25px;
            border-radius: 8px;
            text-align: center;
            margin: 30px 0;
        }

        .success h2 {
            margin-bottom: 10px;
            font-size: 1.5em;
        }

        .footer-links {
            text-align: center;
            padding: 20px;
            background: #f8f9fa;
            border-top: 1px solid #e9ecef;
        }

        .footer-links a {
            color: #6c757d;
            text-decoration: none;
            margin: 0 15px;
            font-size: 0.9em;
        }

        .footer-links a:hover {
            color: #3498db;
        }

        .note {
            background: #e3f2fd;
            border: 1px solid #bbdefb;
            color: #1565c0;
            padding: 12px;
            border-radius: 6px;
            margin: 10px 0;
            font-size: 0.9em;
        }

        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 30px 0;
        }

        .feature {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
        }

        .feature-icon {
            font-size: 2em;
            margin-bottom: 10px;
        }

        .feature h3 {
            color: #2c3e50;
            margin-bottom: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 y-router</h1>
            <p>Claude API 协议转换器 - 高性能 Go 实现</p>
        </div>

        <div class="content">
            <div class="features">
                <div class="feature">
                    <div class="feature-icon">🔄</div>
                    <h3>协议转换</h3>
                    <p>Anthropic Claude API 与 OpenAI 兼容 API 之间的无缝转换</p>
                </div>
                <div class="feature">
                    <div class="feature-icon">🌊</div>
                    <h3>流式支持</h3>
                    <p>完整的流式响应处理能力</p>
                </div>
                <div class="feature">
                    <div class="feature-icon">🚀</div>
                    <h3>高性能</h3>
                    <p>基于 Gin 框架，提供快速响应</p>
                </div>
                <div class="feature">
                    <div class="feature-icon">🔐</div>
                    <h3>安全认证</h3>
                    <p>支持多种 API 密钥认证方式</p>
                </div>
            </div>

            <div class="step">
                <h2><span class="step-number">1</span>环境要求</h2>
                <p>确保您的系统已安装 Go 1.20 或更高版本</p>
                <div class="code-block-wrapper">
                    <div class="code-block">go version</div>
                    <button class="copy-button" onclick="copyToClipboard(this, 'go version')">复制</button>
                </div>
            </div>

            <div class="step">
                <h2><span class="step-number">2</span>安装项目</h2>
                <p>克隆项目并安装依赖</p>
                <div class="code-block-wrapper">
                    <div class="code-block">git clone &lt;repository-url&gt;<br>cd y-router-go<br>go mod tidy</div>
                    <button class="copy-button" onclick="copyToClipboard(this, 'git clone &lt;repository-url&gt;\ncd y-router-go\ngo mod tidy')">复制</button>
                </div>
            </div>

            <div class="step">
                <h2><span class="step-number">3</span>配置环境变量</h2>
                <p>设置必要的环境变量（可选）</p>
                <div class="code-block-wrapper">
                    <div class="code-block"># OpenRouter API 基础 URL（可选）<br>export OPENROUTER_BASE_URL="https://openrouter.ai/api/v1"<br><br># 服务端口（可选，默认 8080）<br>export PORT="8080"</div>
                    <button class="copy-button" onclick="copyToClipboard(this, 'export OPENROUTER_BASE_URL=&quot;https://openrouter.ai/api/v1&quot;\nexport PORT=&quot;8080&quot;')">复制</button>
                </div>
            </div>

            <div class="step">
                <h2><span class="step-number">4</span>运行服务</h2>
                <p>启动 y-router 服务</p>
                <div class="code-block-wrapper">
                    <div class="code-block">go run main.go</div>
                    <button class="copy-button" onclick="copyToClipboard(this, 'go run main.go')">复制</button>
                </div>
                <p>或使用编译后的可执行文件：</p>
                <div class="code-block-wrapper">
                    <div class="code-block">./y-router.exe</div>
                    <button class="copy-button" onclick="copyToClipboard(this, './y-router.exe')">复制</button>
                </div>
            </div>

            <div class="step">
                <h2><span class="step-number">5</span>API 使用示例</h2>
                <p>发送消息请求到 y-router</p>
                <div class="code-block-wrapper">
                    <div class="code-block">curl -X POST http://localhost:8080/v1/messages \n  -H "Content-Type: application/json" \n  -H "Authorization: Bearer YOUR_API_KEY" \n  -d '{\n    "model": "claude-3-sonnet-20240229",\n    "max_tokens": 1024,\n    "messages": [\n      {\n        "role": "user",\n        "content": "Hello, Claude!"\n      }\n    ]\n  }'</div>
                    <button class="copy-button" onclick="copyToClipboard(this, 'curl -X POST http://localhost:8080/v1/messages \\\n  -H \"Content-Type: application/json\" \\\n  -H \"Authorization: Bearer YOUR_API_KEY\" \\\n  -d \'{\\\n    \"model\": \"claude-3-sonnet-20240229\",\\\n    \"max_tokens\": 1024,\\\n    \"messages\": [\\\n      {\\\n        \"role\": \"user\",\\\n        \"content\": \"Hello, Claude!\"\\\n      }\\\n    ]\\\n  }\'')">复制</button>
                </div>
            </div>

            <div class="success">
                <h2>🎉 服务已就绪！</h2>
                <p>y-router 现在正在运行，您可以开始使用 Claude API 协议转换服务</p>
            </div>

            <div class="note">
                <p><strong>API 端点：</strong></p>
                <ul style="margin-top: 10px; margin-left: 20px;">
                    <li><code>POST /v1/messages</code> - 消息处理端点</li>
                    <li><code>GET /</code> - 首页</li>
                    <li><code>GET /terms</code> - 服务条款</li>
                    <li><code>GET /privacy</code> - 隐私政策</li>
                    <li><code>GET /install.sh</code> - 安装脚本</li>
                </ul>
            </div>
        </div>

        <div class="footer-links">
            <a href="https://github.com/luohy15/y-router" target="_blank">项目主页</a>
            <a href="https://openrouter.ai" target="_blank">OpenRouter</a>
            <a href="https://claude.ai/code" target="_blank">Claude Code</a>
            <br>
            <a href="/terms">服务条款</a>
            <a href="/privacy">隐私政策</a>
        </div>
    </div>

    <script>
        function copyToClipboard(button, text) {
            navigator.clipboard.writeText(text).then(function() {
                button.textContent = '已复制！';
                button.classList.add('copied');
                setTimeout(function() {
                    button.textContent = '复制';
                    button.classList.remove('copied');
                }, 2000);
            }).catch(function(err) {
                console.error('复制失败: ', err);
                // 旧浏览器兼容方案
                const textArea = document.createElement('textarea');
                textArea.value = text;
                document.body.appendChild(textArea);
                textArea.focus();
                textArea.select();
                try {
                    document.execCommand('copy');
                    button.textContent = '已复制！';
                    button.classList.add('copied');
                    setTimeout(function() {
                        button.textContent = '复制';
                        button.classList.remove('copied');
                    }, 2000);
                } catch (err) {
                    console.error('兼容方案复制失败', err);
                }
                document.body.removeChild(textArea);
            });
        }
    </script>
</body>
</html>`

const termsHtml = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>服务条款 - y-router</title>
    <link rel="shortcut icon" type="image/svg+xml" href="%s">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'PingFang SC', 'Microsoft YaHei', sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f8f9fa;
            padding: 20px;
        }

        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
            padding: 40px;
        }

        h1 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 2.5em;
            font-weight: 300;
        }

        .last-updated {
            color: #6c757d;
            margin-bottom: 30px;
            font-size: 0.9em;
        }

        h2 {
            color: #34495e;
            margin-top: 30px;
            margin-bottom: 15px;
            font-size: 1.5em;
        }

        p {
            margin-bottom: 15px;
        }

        ul {
            margin-bottom: 20px;
            padding-left: 20px;
        }

        li {
            margin-bottom: 8px;
        }

        .back-link {
            display: inline-block;
            margin-top: 30px;
            color: #3498db;
            text-decoration: none;
            font-weight: 500;
        }

        .back-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>服务条款</h1>
        <div class="last-updated">最后更新：2025年11月24日</div>

        <h2>1. 接受条款</h2>
        <p>通过访问和使用 y-router 服务（"服务"），您接受并同意受本协议条款和规定的约束。</p>

        <h2>2. 服务描述</h2>
        <p>y-router 是一个协议转换服务，旨在实现 Anthropic Claude API 与 OpenAI 兼容 API 之间的兼容性。该服务作为中介，在不同 API 格式之间转换请求和响应。</p>

        <h2>3. 用户责任</h2>
        <p>用户有责任：</p>
        <ul>
            <li>维护其 API 密钥的安全性</li>
            <li>确保遵守上游 API 提供商的服务条款</li>
            <li>根据适用法律和法规使用服务</li>
            <li>不尝试规避速率限制或其他使用限制</li>
        </ul>

        <h2>4. 隐私和数据</h2>
        <p>有关我们如何收集、使用和保护您的数据的信息，请参阅我们的隐私政策。</p>

        <h2>5. 服务可用性</h2>
        <p>我们努力维持服务的高可用性，但不保证不间断的访问。服务可能因维护、更新或其他原因而暂时不可用。</p>

        <h2>6. 责任限制</h2>
        <p>服务按"原样"提供，不作任何形式的保证。对于因您使用服务而产生的任何间接、附带、特殊或后果性损害，我们概不负责。</p>

        <h2>7. 条款变更</h2>
        <p>我们保留随时修改这些条款的权利。变更将在发布后立即生效。您继续使用服务即表示接受任何修改后的条款。</p>

        <h2>8. 联系方式</h2>
        <p>如果您对这些服务条款有疑问，请通过项目仓库联系我们。</p>

        <a href="/" class="back-link">← 返回首页</a>
    </div>
</body>
</html>`

const privacyHtml = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>隐私政策 - y-router</title>
    <link rel="shortcut icon" type="image/svg+xml" href="%s">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'PingFang SC', 'Microsoft YaHei', sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f8f9fa;
            padding: 20px;
        }

        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
            padding: 40px;
        }

        h1 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 2.5em;
            font-weight: 300;
        }

        .last-updated {
            color: #6c757d;
            margin-bottom: 30px;
            font-size: 0.9em;
        }

        h2 {
            color: #34495e;
            margin-top: 30px;
            margin-bottom: 15px;
            font-size: 1.5em;
        }

        p {
            margin-bottom: 15px;
        }

        ul {
            margin-bottom: 20px;
            padding-left: 20px;
        }

        li {
            margin-bottom: 8px;
        }

        .back-link {
            display: inline-block;
            margin-top: 30px;
            color: #3498db;
            text-decoration: none;
            font-weight: 500;
        }

        .back-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>隐私政策</h1>
        <div class="last-updated">最后更新：2025年11月24日</div>

        <h2>1. 我们收集的信息</h2>
        <p>y-router 设计为最小化数据收集。我们可能收集：</p>
        <ul>
            <li>API 请求元数据（时间戳、模型名称、令牌使用情况）</li>
            <li>用于故障排除和服务改进的错误日志</li>
            <li>了解服务性能的基本使用分析</li>
        </ul>

        <h2>2. 我们不收集的信息</h2>
        <p>我们不收集：</p>
        <ul>
            <li>您的 API 请求或响应的内容</li>
            <li>您的 API 密钥或身份验证凭据</li>
            <li>个人身份识别信息</li>
            <li>对话内容或聊天历史记录</li>
        </ul>

        <h2>3. 数据处理</h2>
        <p>该服务作为协议转换器，处理传输中的数据。除了转换过程所需的内容外，我们不存储对话内容或 API 请求负载。</p>

        <h2>4. 数据共享</h2>
        <p>我们不会出于营销目的出售、出租或与第三方共享您的数据。您的 API 请求会根据服务运行需要转发给上游提供商。</p>

        <h2>5. 数据安全</h2>
        <p>我们实施合理的安全措施来保护我们处理的数据。但是，没有通过互联网传输的方法是 100% 安全的。</p>

        <h2>6. Cookie 和跟踪</h2>
        <p>对于核心 API 服务，我们不使用 Cookie 或跟踪技术。网站可能会使用基本分析来了解使用模式。</p>

        <h2>7. 您的权利</h2>
        <p>您有权：</p>
        <ul>
            <li>访问我们拥有的关于您的任何数据</li>
            <li>请求删除您的数据</li>
            <li>在技术上可行的情况下选择退出数据收集</li>
        </ul>

        <h2>8. 隐私政策变更</h2>
        <p>我们可能会不时更新此隐私政策。变更将在此页面上发布并附上更新日期。</p>

        <h2>9. 联系方式</h2>
        <p>如果您对此隐私政策有疑问，请通过项目仓库联系我们。</p>

        <a href="/" class="back-link">← 返回首页</a>
    </div>
</body>
</html>`

const installSh = `#!/bin/bash

set -e

install_nodejs() {
    local platform=$(uname -s)
    
    case "$platform" in
        Linux|Darwin)
            echo "🚀 Installing Node.js on Unix/Linux/macOS..."
            
            echo "📥 Downloading and installing nvm..."
            curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
            
            echo "🔄 Loading nvm environment..."
            \. "$HOME/.nvm/nvm.sh"
            
            echo "📦 Downloading and installing Node.js v22..."
            nvm install 22
            
            echo -n "✅ Node.js installation completed! Version: "
            node -v # Should print "v22.17.0".
            echo -n "✅ Current nvm version: "
            nvm current # Should print "v22.17.0".
            echo -n "✅ npm version: "
            npm -v # Should print "10.9.2".
            ;;
        *)
            echo "Unsupported platform: $platform"
            exit 1
            ;;
    esac
}

# Check if Node.js is already installed and version is >= 18
if command -v node >/dev/null 2>&1; then
    current_version=$(node -v | sed 's/v//')
    major_version=$(echo $current_version | cut -d. -f1)
    
    if [ "$major_version" -ge 18 ]; then
        echo "Node.js is already installed: v$current_version"
    else
        echo "Node.js v$current_version is installed but version < 18. Upgrading..."
        install_nodejs
    fi
else
    echo "Node.js not found. Installing..."
    install_nodejs
fi

echo "🔧 Installing Claude Code..."
npm install -g @anthropic-ai/claude-code

echo "📝 Setting up environment variables..."
read -p "Enter your OpenRouter API key (or press Enter to use Moonshot): " api_key

if [ -z "$api_key" ]; then
    echo "Using Moonshot as default provider..."
    api_key="sk-moonshot-key-placeholder"
    base_url="https://cc.yovy.app"
else
    base_url="https://cc.yovy.app"
fi

# Detect shell and update appropriate config file
if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
    shell_config="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ] || [ -f "$HOME/.bashrc" ]; then
    shell_config="$HOME/.bashrc"
else
    shell_config="$HOME/.profile"
fi

echo "" >> "$shell_config"
echo "# Claude Code configuration" >> "$shell_config"
echo "export ANTHROPIC_BASE_URL=\"$base_url\"" >> "$shell_config"
echo "export ANTHROPIC_API_KEY=\"$api_key\"" >> "$shell_config"

echo "🎉 Installation completed!"
echo "Please restart your terminal or run: source $shell_config"
echo "Then you can start using Claude Code by typing: claude"

# Optional: Ask about model configuration
read -p "Do you want to configure specific models? (y/N): " configure_models

if [[ $configure_models =~ ^[Yy]$ ]]; then
    echo "Available models:"
    echo "- moonshotai/kimi-k2 (recommended)"
    echo "- google/gemini-2.5-flash"
    echo "- anthropic/claude-3.5-sonnet"
    
    read -p "Enter your preferred model (or press Enter for default): " preferred_model
    
    if [ -n "$preferred_model" ]; then
        echo "export ANTHROPIC_MODEL=\"$preferred_model\"" >> "$shell_config"
    fi
    
    read -p "Enter small/fast model (or press Enter for default): " small_model
    
    if [ -n "$small_model" ]; then
        echo "export ANTHROPIC_SMALL_FAST_MODEL=\"$small_model\"" >> "$shell_config"
    fi
fi

echo "✨ Setup complete! Restart your terminal and run 'claude' to start using Claude Code."
`

// handleIndex 处理首页请求
func handleIndex(c *gin.Context) {
	faviconUrl := generateFaviconDataUrl()
	html := fmt.Sprintf(indexHtml, faviconUrl)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// handleTerms 处理服务条款页面请求
func handleTerms(c *gin.Context) {
	faviconUrl := generateFaviconDataUrl()
	html := fmt.Sprintf(termsHtml, faviconUrl)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// handlePrivacy 处理隐私政策页面请求
func handlePrivacy(c *gin.Context) {
	faviconUrl := generateFaviconDataUrl()
	html := fmt.Sprintf(privacyHtml, faviconUrl)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// handleInstallSh 处理安装脚本请求
func handleInstallSh(c *gin.Context) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, installSh)
}