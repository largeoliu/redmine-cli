
# 项目编码检测和修复脚本
# 用于检测和修复项目中的文件编码问题

param(
    [switch]$Fix,
    [string]$Path = "."
)

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  Redmine CLI 编码检测和修复工具" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# 检测文件是否有 BOM
function Test-UTF8BOM {
    param([string]$FilePath)
    
    $bytes = [System.IO.File]::ReadAllBytes($FilePath)
    if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) {
        return $true
    }
    return $false
}

# 检测文件是否包含乱码字符
function Test-ContainsGarbled {
    param([string]$FilePath)
    
    try {
        $content = [System.IO.File]::ReadAllText($FilePath, [System.Text.Encoding]::UTF8)
        if ($content -match "[\uFFFD]") {
            return $true
        }
        return $false
    }
    catch {
        return $true
    }
}

# 获取文件扩展名
$extensions = @("*.md", "*.go", "*.txt", "*.json", "*.yaml", "*.yml", "*.html", "*.css", "*.js", "*.ts")

Write-Host "正在扫描文件..." -ForegroundColor Yellow
Write-Host ""

$filesWithIssues = @()

foreach ($ext in $extensions) {
    $files = Get-ChildItem -Path $Path -Filter $ext -Recurse -File | 
               Where-Object { $_.FullName -notmatch "[\\/](\.git|node_modules|coverage|bin|obj)[\\/]" }
    
    foreach ($file in $files) {
        $hasBOM = Test-UTF8BOM -FilePath $file.FullName
        $hasGarbled = Test-ContainsGarbled -FilePath $file.FullName
        
        if ($hasBOM -or $hasGarbled) {
            $issue = ""
            if ($hasBOM) { $issue += "有 UTF-8 BOM; " }
            if ($hasGarbled) { $issue += "包含乱码字符" }
            
            $filesWithIssues += @{
                Path = $file.FullName
                Issue = $issue.Trim()
                HasBOM = $hasBOM
                HasGarbled = $hasGarbled
            }
            
            Write-Host "  [问题] $($file.FullName)" -ForegroundColor Red
            Write-Host "         问题: $issue" -ForegroundColor DarkRed
        }
    }
}

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  扫描完成" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

if ($filesWithIssues.Count -eq 0) {
    Write-Host "✓ 没有发现编码问题！" -ForegroundColor Green
    exit 0
}

Write-Host "发现 $($filesWithIssues.Count) 个文件有编码问题" -ForegroundColor Yellow
Write-Host ""

if ($Fix) {
    Write-Host "=========================================" -ForegroundColor Cyan
    Write-Host "  开始修复..." -ForegroundColor Cyan
    Write-Host "=========================================" -ForegroundColor Cyan
    Write-Host ""
    
    $utf8NoBom = New-Object System.Text.UTF8Encoding $false
    $fixedCount = 0
    
    foreach ($fileInfo in $filesWithIssues) {
        try {
            Write-Host "  修复: $($fileInfo.Path)" -ForegroundColor Yellow
            
            # 读取文件内容，先用 UTF-8 尝试
            $content = [System.IO.File]::ReadAllText($fileInfo.Path, [System.Text.Encoding]::UTF8)
            
            # 写回无 BOM 的 UTF-8
            [System.IO.File]::WriteAllText($fileInfo.Path, $content, $utf8NoBom)
            
            Write-Host "    ✓ 修复完成" -ForegroundColor Green
            $fixedCount++
        }
        catch {
            Write-Host "    ✗ 修复失败: $_" -ForegroundColor Red
        }
    }
    
    Write-Host ""
    Write-Host "=========================================" -ForegroundColor Cyan
    Write-Host "  修复完成" -ForegroundColor Cyan
    Write-Host "=========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "成功修复 $fixedCount / $($filesWithIssues.Count) 个文件" -ForegroundColor Green
}
else {
    Write-Host "使用 -Fix 参数来自动修复这些问题" -ForegroundColor Cyan
    Write-Host "示例: .\scripts\check-encoding.ps1 -Fix" -ForegroundColor Cyan
}

