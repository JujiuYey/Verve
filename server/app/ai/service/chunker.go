package service

import (
	"fmt"
	"strings"
)

// ChunkType represents the type of semantic unit
type ChunkType int

const (
	ChunkTypeParagraph ChunkType = iota
	ChunkTypeHeading
	ChunkTypeTable
	ChunkTypeCodeBlock
	ChunkTypeList
	ChunkTypeASCIIArt
)

// ChunkerConfig holds configuration for the chunker
type ChunkerConfig struct {
	MaxChunkSize     int  // 最大 chunk 字符数
	MinChunkSize     int  // 最小 chunk 字符数（过小会合并）
	PreserveASCIIArt bool // 保留 ASCII 图表
	PreserveTables   bool // 保留表格
}

// DefaultChunkerConfig returns sensible defaults
func DefaultChunkerConfig() *ChunkerConfig {
	return &ChunkerConfig{
		MaxChunkSize:     800, // 增大默认值，提升语义完整性
		MinChunkSize:     200, // 增大最小值，减少碎片
		PreserveASCIIArt: true,
		PreserveTables:   true,
	}
}

type Chunker struct {
	config *ChunkerConfig
}

func NewChunker(maxChunkSize, minChunkSize int) *Chunker {
	return &Chunker{
		config: &ChunkerConfig{
			MaxChunkSize: maxChunkSize,
			MinChunkSize: minChunkSize,
		},
	}
}

// NewSemanticChunker creates a chunker with semantic awareness
func NewSemanticChunker(config *ChunkerConfig) *Chunker {
	if config == nil {
		config = DefaultChunkerConfig()
	}
	return &Chunker{config: config}
}

// Chunk represents a semantic text chunk
type Chunk struct {
	Index       int
	Text        string
	Size        int       // 字符数
	Type        ChunkType // chunk 类型
	SectionPath string    // 所属章节路径
}

// ChunkText 对文本进行语义分块
func (c *Chunker) ChunkText(text string) []*Chunk {
	if text == "" {
		return nil
	}

	// 清理无效 UTF-8 字符
	text = strings.ToValidUTF8(text, "")

	// 1. 识别并提取特殊结构
	sections := c.parseDocument(text)

	// 2. 对每个语义单元进行分块
	chunks := c.processSections(sections)

	// 3. 合并过短的 chunks
	chunks = c.mergeSmallChunks(chunks)

	// 4. 重新编号
	for i, chunk := range chunks {
		chunk.Index = i
	}

	return chunks
}

// Section represents a parsed document section
type Section struct {
	Type     ChunkType
	Content  string
	Level    int      // 标题级别（1-6）
	RawLines []string // 原始行
}

// parseDocument 解析文档结构
func (c *Chunker) parseDocument(text string) []*Section {
	lines := strings.Split(text, "\n")
	sections := make([]*Section, 0)

	var currentParagraph []string

	flushParagraph := func() {
		if len(currentParagraph) > 0 {
			content := strings.TrimSpace(strings.Join(currentParagraph, "\n"))
			if content != "" {
				sections = append(sections, &Section{
					Type:     ChunkTypeParagraph,
					Content:  content,
					RawLines: currentParagraph,
				})
			}
			currentParagraph = nil
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// 检测 ASCII 艺术图表（Box Drawing 字符）
		if c.isASCIIArtLine(line) {
			flushParagraph()
			// 收集整个 ASCII 图表
			var artLines []string
			for i < len(lines) && c.isASCIIArtLine(lines[i]) {
				artLines = append(artLines, lines[i])
				i++
			}
			i-- // 回退一步

			if c.config.PreserveASCIIArt {
				sections = append(sections, &Section{
					Type:     ChunkTypeASCIIArt,
					Content:  strings.Join(artLines, "\n"),
					RawLines: artLines,
				})
			}
			continue
		}

		// 检测代码块
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			flushParagraph()
			var codeLines []string
			codeLines = append(codeLines, line)
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				codeLines = append(codeLines, lines[i])
				i++
			}
			if i < len(lines) {
				codeLines = append(codeLines, lines[i]) // 结束标记
			}

			sections = append(sections, &Section{
				Type:     ChunkTypeCodeBlock,
				Content:  strings.Join(codeLines, "\n"),
				RawLines: codeLines,
			})
			continue
		}

		// 检测 Markdown 标题
		if matched, level := c.isHeading(line); matched {
			flushParagraph()
			headingText := strings.TrimSpace(strings.TrimPrefix(line, strings.Repeat("#", level)))
			// 去除可能的空格
			headingText = strings.TrimPrefix(headingText, " ")
			sections = append(sections, &Section{
				Type:     ChunkTypeHeading,
				Level:    level,
				Content:  headingText,
				RawLines: []string{line},
			})
			continue
		}

		// 检测表格
		if c.isTableRow(line) {
			flushParagraph()
			var tableLines []string
			// 收集所有连续的表格行
			for i < len(lines) && c.isTableRow(lines[i]) {
				tableLines = append(tableLines, lines[i])
				i++
			}
			i-- // 回退一步

			if c.config.PreserveTables {
				sections = append(sections, &Section{
					Type:     ChunkTypeTable,
					Content:  strings.Join(tableLines, "\n"),
					RawLines: tableLines,
				})
			}
			continue
		}

		// 其他内容作为段落
		if trimmed != "" {
			currentParagraph = append(currentParagraph, line)
		} else if len(currentParagraph) > 0 {
			flushParagraph()
		}
	}

	flushParagraph()
	return sections
}

// isASCIIArtLine 检测是否是 ASCII 艺术图表行
func (c *Chunker) isASCIIArtLine(line string) bool {
	boxChars := "─│┌┐└┘├┤┬┴┼━┃┏┓┗┛┣┫┳┻╋"
	count := 0
	for _, ch := range line {
		if strings.Contains(boxChars, string(ch)) {
			count++
		}
	}
	return count >= 3
}

// isHeading 检测 Markdown 标题
func (c *Chunker) isHeading(line string) (bool, int) {
	trimmed := strings.TrimSpace(line)
	for i := 1; i <= 6; i++ {
		prefix := strings.Repeat("#", i) + " "
		if strings.HasPrefix(trimmed, prefix) {
			return true, i
		}
	}
	return false, 0
}

// isTableRow 检测 Markdown 表格行
// 严格的表格检测：必须有 | 开头和结尾，且包含有效的单元格
func (c *Chunker) isTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	
	// 必须以 | 开头和结尾
	if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
		return false
	}
	
	// 分隔符行（如 |---|---| 或 |:---|:---|）直接返回 true
	if strings.Contains(trimmed, "|--") || strings.Contains(trimmed, "|-") {
		return true
	}
	
	// 内容行：分析单元格内容
	// 去除首尾的 |
	content := trimmed[1 : len(trimmed)-1]
	cols := strings.Split(content, "|")
	
	// 至少需要 2 个非空列（表头行）
	validCols := 0
	for _, col := range cols {
		trimmedCol := strings.TrimSpace(col)
		// 排除纯分隔符内容
		if len(trimmedCol) > 0 && !strings.HasPrefix(trimmedCol, "-") {
			validCols++
		}
	}
	
	return validCols >= 2
}

// processSections 处理所有语义单元
func (c *Chunker) processSections(sections []*Section) []*Chunk {
	chunks := make([]*Chunk, 0)
	var currentChunk *Chunk

	for _, section := range sections {
		switch section.Type {
		case ChunkTypeASCIIArt, ChunkTypeCodeBlock, ChunkTypeTable:
			// 先保存当前的段落 chunk
			if currentChunk != nil {
				if currentChunk.Size >= c.config.MinChunkSize {
					chunks = append(chunks, currentChunk)
				} else {
					section.Content = currentChunk.Text + "\n\n" + section.Content
				}
				currentChunk = nil
			}

			if len(section.Content) > 0 {
				chunks = append(chunks, &Chunk{
					Index: len(chunks),
					Text:  section.Content,
					Size:  len(section.Content),
					Type:  section.Type,
				})
			}

		case ChunkTypeHeading:
			if currentChunk != nil {
				if currentChunk.Size >= c.config.MinChunkSize {
					chunks = append(chunks, currentChunk)
				}
				currentChunk = nil
			}
			if len(section.Content) > 0 {
				chunks = append(chunks, &Chunk{
					Index: len(chunks),
					Text:  section.Content,
					Size:  len(section.Content),
					Type:  ChunkTypeHeading,
				})
			}

		case ChunkTypeParagraph:
			if currentChunk == nil {
				currentChunk = &Chunk{
					Text: section.Content,
					Size: len(section.Content),
					Type: ChunkTypeParagraph,
				}
			} else if currentChunk.Size+len(section.Content) <= c.config.MaxChunkSize {
				currentChunk.Text += "\n\n" + section.Content
				currentChunk.Size += 2 + len(section.Content)
			} else {
				if currentChunk.Size >= c.config.MinChunkSize {
					chunks = append(chunks, currentChunk)
				}
				currentChunk = &Chunk{
					Text: section.Content,
					Size: len(section.Content),
					Type: ChunkTypeParagraph,
				}
			}
		}
	}

	if currentChunk != nil && currentChunk.Size >= c.config.MinChunkSize {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// mergeSmallChunks 合并过短的相邻 chunks
func (c *Chunker) mergeSmallChunks(chunks []*Chunk) []*Chunk {
	if len(chunks) == 0 {
		return chunks
	}

	result := make([]*Chunk, 0, len(chunks))
	current := chunks[0]

	for i := 1; i < len(chunks); i++ {
		chunk := chunks[i]
		if chunk.Type == ChunkTypeASCIIArt || chunk.Type == ChunkTypeCodeBlock || chunk.Type == ChunkTypeTable {
			if current.Size >= c.config.MinChunkSize {
				result = append(result, current)
			} else {
				chunk.Text = current.Text + "\n\n" + chunk.Text
				chunk.Size = len(chunk.Text)
			}
			current = chunk
			continue
		}

		if current.Size < c.config.MinChunkSize {
			current.Text += "\n" + chunk.Text
			current.Size += 1 + chunk.Size
		} else {
			result = append(result, current)
			current = chunk
		}
	}

	if current != nil {
		result = append(result, current)
	}

	return result
}

// ValidateChunkSize 验证向量维度是否符合预期
func (c *Chunker) ValidateChunkSize(vectorDim int) error {
	if vectorDim != 1024 && vectorDim != 2048 {
		return fmt.Errorf("不支持的向量维度: %d", vectorDim)
	}
	return nil
}

// ChunkTypeString returns string representation of chunk type
func (ct ChunkType) String() string {
	switch ct {
	case ChunkTypeParagraph:
		return "paragraph"
	case ChunkTypeHeading:
		return "heading"
	case ChunkTypeTable:
		return "table"
	case ChunkTypeCodeBlock:
		return "code"
	case ChunkTypeList:
		return "list"
	case ChunkTypeASCIIArt:
		return "ascii_art"
	default:
		return "unknown"
	}
}
