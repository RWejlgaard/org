package ui

import "strings"

// renderLatexMath converts LaTeX math expressions to Unicode for terminal display
func renderLatexMath(latex string) string {
	result := latex

	// Remove LaTeX math delimiters
	result = strings.ReplaceAll(result, `\(`, "")
	result = strings.ReplaceAll(result, `\)`, "")
	result = strings.ReplaceAll(result, `\[`, "")
	result = strings.ReplaceAll(result, `\]`, "")
	result = strings.ReplaceAll(result, `$$`, "")

	// Remove single $ delimiters (being careful not to remove actual dollar signs in text)
	// Simple approach: if line starts/ends with $, remove it
	result = strings.TrimSpace(result)
	if strings.HasPrefix(result, "$") && strings.HasSuffix(result, "$") && len(result) > 2 {
		result = result[1 : len(result)-1]
		result = strings.TrimSpace(result)
	}

	// Map of common LaTeX commands to Unicode equivalents
	replacements := map[string]string{
		// Greek letters (lowercase)
		`\alpha`:   "α",
		`\beta`:    "β",
		`\gamma`:   "γ",
		`\delta`:   "δ",
		`\epsilon`: "ε",
		`\zeta`:    "ζ",
		`\eta`:     "η",
		`\theta`:   "θ",
		`\iota`:    "ι",
		`\kappa`:   "κ",
		`\lambda`:  "λ",
		`\mu`:      "μ",
		`\nu`:      "ν",
		`\xi`:      "ξ",
		`\pi`:      "π",
		`\rho`:     "ρ",
		`\sigma`:   "σ",
		`\tau`:     "τ",
		`\upsilon`: "υ",
		`\phi`:     "φ",
		`\chi`:     "χ",
		`\psi`:     "ψ",
		`\omega`:   "ω",

		// Greek letters (uppercase)
		`\Gamma`:   "Γ",
		`\Delta`:   "Δ",
		`\Theta`:   "Θ",
		`\Lambda`:  "Λ",
		`\Xi`:      "Ξ",
		`\Pi`:      "Π",
		`\Sigma`:   "Σ",
		`\Upsilon`: "Υ",
		`\Phi`:     "Φ",
		`\Psi`:     "Ψ",
		`\Omega`:   "Ω",

		// Math operators
		`\times`:   "×",
		`\div`:     "÷",
		`\pm`:      "±",
		`\mp`:      "∓",
		`\cdot`:    "·",
		`\star`:    "⋆",
		`\ast`:     "∗",
		`\circ`:    "∘",
		`\bullet`:  "•",

		// Relations
		`\le`:      "≤",
		`\ge`:      "≥",
		`\leq`:     "≤",
		`\geq`:     "≥",
		`\ne`:      "≠",
		`\neq`:     "≠",
		`\approx`:  "≈",
		`\equiv`:   "≡",
		`\sim`:     "∼",
		`\simeq`:   "≃",
		`\propto`:  "∝",

		// Arrows
		`\to`:         "→",
		`\rightarrow`: "→",
		`\leftarrow`:  "←",
		`\Rightarrow`: "⇒",
		`\Leftarrow`:  "⇐",
		`\mapsto`:     "↦",

		// Set theory
		`\in`:       "∈",
		`\notin`:    "∉",
		`\subset`:   "⊂",
		`\supset`:   "⊃",
		`\subseteq`: "⊆",
		`\supseteq`: "⊇",
		`\cup`:      "∪",
		`\cap`:      "∩",
		`\emptyset`: "∅",
		`\forall`:   "∀",
		`\exists`:   "∃",

		// Calculus
		`\partial`: "∂",
		`\nabla`:   "∇",
		`\int`:     "∫",
		`\sum`:     "∑",
		`\prod`:    "∏",
		`\infty`:   "∞",

		// Logic
		`\land`:  "∧",
		`\lor`:   "∨",
		`\lnot`:  "¬",
		`\neg`:   "¬",
		`\wedge`: "∧",
		`\vee`:   "∨",

		// Special symbols
		`\hbar`:     "ℏ",
		`\ell`:      "ℓ",
		`\Re`:       "ℜ",
		`\Im`:       "ℑ",
		`\angle`:    "∠",
		`\triangle`: "△",
		`\square`:   "□",
		`\degree`:   "°",

		// Superscripts (common ones)
		`^0`: "⁰",
		`^1`: "¹",
		`^2`: "²",
		`^3`: "³",
		`^4`: "⁴",
		`^5`: "⁵",
		`^6`: "⁶",
		`^7`: "⁷",
		`^8`: "⁸",
		`^9`: "⁹",
		`^+`: "⁺",
		`^-`: "⁻",
		`^=`: "⁼",
		`^(`: "⁽",
		`^)`: "⁾",

		// Subscripts (common ones)
		`_0`: "₀",
		`_1`: "₁",
		`_2`: "₂",
		`_3`: "₃",
		`_4`: "₄",
		`_5`: "₅",
		`_6`: "₆",
		`_7`: "₇",
		`_8`: "₈",
		`_9`: "₉",
		`_+`: "₊",
		`_-`: "₋",
		`_=`: "₌",
		`_(`: "₍",
		`_)`: "₎",
	}

	// Apply all replacements
	for latexCmd, unicode := range replacements {
		result = strings.ReplaceAll(result, latexCmd, unicode)
	}

	// Handle simple fractions \frac{a}{b} -> a/b
	result = handleFractions(result)

	// Handle square roots \sqrt{x} -> √(x)
	result = handleSquareRoots(result)

	// Handle superscripts ^{...} and subscripts _{...}
	result = handleSuperscripts(result)
	result = handleSubscripts(result)

	return result
}

// handleFractions converts \frac{numerator}{denominator} to numerator/denominator
func handleFractions(text string) string {
	// Simple regex-free approach for basic fractions
	result := text
	for {
		start := strings.Index(result, `\frac{`)
		if start == -1 {
			break
		}

		// Find the numerator
		numStart := start + 6
		numEnd, numerator := findBracedContent(result, numStart)
		if numEnd == -1 {
			break
		}

		// Find the denominator
		if numEnd >= len(result) || result[numEnd] != '{' {
			break
		}
		denomEnd, denominator := findBracedContent(result, numEnd+1)
		if denomEnd == -1 {
			break
		}

		// Replace \frac{num}{denom} with (num)/(denom)
		replacement := "(" + numerator + ")/(" + denominator + ")"
		result = result[:start] + replacement + result[denomEnd:]
	}
	return result
}

// handleSquareRoots converts \sqrt{x} to √(x)
func handleSquareRoots(text string) string {
	result := text
	for {
		start := strings.Index(result, `\sqrt{`)
		if start == -1 {
			break
		}

		// Find the content
		contentStart := start + 6
		contentEnd, content := findBracedContent(result, contentStart)
		if contentEnd == -1 {
			break
		}

		// Replace \sqrt{content} with √(content)
		replacement := "√(" + content + ")"
		result = result[:start] + replacement + result[contentEnd:]
	}
	return result
}

// findBracedContent finds content within braces starting at position i
// Returns the position after the closing brace and the content
func findBracedContent(text string, start int) (int, string) {
	if start >= len(text) {
		return -1, ""
	}

	depth := 1
	i := start

	for i < len(text) && depth > 0 {
		if text[i] == '{' {
			depth++
		} else if text[i] == '}' {
			depth--
			if depth == 0 {
				return i + 1, text[start:i]
			}
		}
		i++
	}

	return -1, ""
}

// handleSuperscripts converts ^{...} to Unicode superscripts where possible
func handleSuperscripts(text string) string {
	superscriptMap := map[rune]string{
		'0': "⁰", '1': "¹", '2': "²", '3': "³", '4': "⁴",
		'5': "⁵", '6': "⁶", '7': "⁷", '8': "⁸", '9': "⁹",
		'a': "ᵃ", 'b': "ᵇ", 'c': "ᶜ", 'd': "ᵈ", 'e': "ᵉ",
		'f': "ᶠ", 'g': "ᵍ", 'h': "ʰ", 'i': "ⁱ", 'j': "ʲ",
		'k': "ᵏ", 'l': "ˡ", 'm': "ᵐ", 'n': "ⁿ", 'o': "ᵒ",
		'p': "ᵖ", 'r': "ʳ", 's': "ˢ", 't': "ᵗ", 'u': "ᵘ",
		'v': "ᵛ", 'w': "ʷ", 'x': "ˣ", 'y': "ʸ", 'z': "ᶻ",
		'A': "ᴬ", 'B': "ᴮ", 'D': "ᴰ", 'E': "ᴱ", 'G': "ᴳ",
		'H': "ᴴ", 'I': "ᴵ", 'J': "ᴶ", 'K': "ᴷ", 'L': "ᴸ",
		'M': "ᴹ", 'N': "ᴺ", 'O': "ᴼ", 'P': "ᴾ", 'R': "ᴿ",
		'T': "ᵀ", 'U': "ᵁ", 'V': "ⱽ", 'W': "ᵂ",
		'+': "⁺", '-': "⁻", '=': "⁼", '(': "⁽", ')': "⁾",
	}

	result := text

	// Handle ^{...} format
	for {
		start := strings.Index(result, "^{")
		if start == -1 {
			break
		}

		contentStart := start + 2
		contentEnd, content := findBracedContent(result, contentStart)
		if contentEnd == -1 {
			break
		}

		// Convert content to superscript
		var superscript strings.Builder
		for _, ch := range content {
			if sup, ok := superscriptMap[ch]; ok {
				superscript.WriteString(sup)
			} else {
				// If no superscript version exists, wrap in parentheses
				superscript.WriteRune(ch)
			}
		}

		result = result[:start] + superscript.String() + result[contentEnd:]
	}

	// Handle simple ^x format (single character without braces)
	for i := 0; i < len(result)-1; i++ {
		if result[i] == '^' && result[i+1] != '{' {
			ch := rune(result[i+1])
			if sup, ok := superscriptMap[ch]; ok {
				result = result[:i] + sup + result[i+2:]
			}
		}
	}

	return result
}

// handleSubscripts converts _{...} to Unicode subscripts where possible
func handleSubscripts(text string) string {
	subscriptMap := map[rune]string{
		'0': "₀", '1': "₁", '2': "₂", '3': "₃", '4': "₄",
		'5': "₅", '6': "₆", '7': "₇", '8': "₈", '9': "₉",
		'a': "ₐ", 'e': "ₑ", 'h': "ₕ", 'i': "ᵢ", 'j': "ⱼ",
		'k': "ₖ", 'l': "ₗ", 'm': "ₘ", 'n': "ₙ", 'o': "ₒ",
		'p': "ₚ", 'r': "ᵣ", 's': "ₛ", 't': "ₜ", 'u': "ᵤ",
		'v': "ᵥ", 'x': "ₓ",
		'+': "₊", '-': "₋", '=': "₌", '(': "₍", ')': "₎",
	}

	result := text

	// Handle _{...} format
	for {
		start := strings.Index(result, "_{")
		if start == -1 {
			break
		}

		contentStart := start + 2
		contentEnd, content := findBracedContent(result, contentStart)
		if contentEnd == -1 {
			break
		}

		// Convert content to subscript
		var subscript strings.Builder
		for _, ch := range content {
			if sub, ok := subscriptMap[ch]; ok {
				subscript.WriteString(sub)
			} else {
				// If no subscript version exists, wrap in parentheses
				subscript.WriteRune(ch)
			}
		}

		result = result[:start] + subscript.String() + result[contentEnd:]
	}

	// Handle simple _x format (single character without braces)
	for i := 0; i < len(result)-1; i++ {
		if result[i] == '_' && result[i+1] != '{' {
			ch := rune(result[i+1])
			if sub, ok := subscriptMap[ch]; ok {
				result = result[:i] + sub + result[i+2:]
			}
		}
	}

	return result
}
