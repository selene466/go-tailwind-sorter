package config

import (
	"fmt"
	"os"
	"sort"

	"github.com/BurntSushi/toml"
)

const DefaultConfigFileName string = "tailwind-sorter.toml"

type TomlRoot struct {
	Tool ToolSelection `toml:"tool"`
}

type ToolSelection struct {
	Sorter UserConfig `toml:"tailwind_sorter"`
}

type UserConfig struct {
	FilePatterns    []string `toml:"file_patterns"`
	ClassAttributes []string `toml:"class_attributes"`
}

type Config struct {
	ClassOrder      []string
	VariantOrder    map[string]int
	FilePatterns    []string
	ClassAttributes []string
}

func New(configFile string) (*Config, error) {
	config := defaultConfig()

	if configFile == "" {
		if _, err := os.Stat(DefaultConfigFileName); err == nil {
			configFile = DefaultConfigFileName
		}
	}

	if configFile != "" {
		var tomlRoot TomlRoot
		if _, err := toml.DecodeFile(configFile, &tomlRoot); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
		}

		config.merge(&tomlRoot.Tool.Sorter)
	}

	return config, nil
}

func defaultConfig() *Config {
	return &Config{
		ClassOrder: []string{
			// Accessibility
			"sr-only", "not-sr-only",

			// Core / Layout
			"container", "columns-", "break-after-", "break-before-", "break-inside-",
			"box-decoration-", "box-border", "box-content",
			"block", "inline-block", "inline", "flex", "inline-flex", "table", "inline-table",
			"table-caption", "table-cell", "table-column", "table-column-group", "table-footer-group",
			"table-header-group", "table-row-group", "table-row", "flow-root", "grid", "inline-grid",
			"contents", "list-item", "hidden",
			"float-", "clear-", "isolate", "isolation-auto",
			"object-", "overflow-", "overscroll-",
			"static", "fixed", "absolute", "relative", "sticky",
			"inset-", "top-", "right-", "bottom-", "left-",
			"visible", "invisible", "collapse", "z-",

			// Flexbox & Grid Container
			"basis-", "flex-row", "flex-col", "flex-wrap", "flex-nowrap", "flex-",
			"grow", "shrink", "flex-grow", "flex-shrink", "order-",
			"grid-cols-", "col-", "grid-rows-", "row-", "grid-flow-", "auto-cols-", "auto-rows-",
			"gap-", "justify-", "content-", "items-", "self-", "place-",

			// Sizing (v4 size- first)
			"size-", "w-", "min-w-", "max-w-", "h-", "min-h-", "max-h-", "field-sizing-",

			// Spacing
			"p-", "px-", "py-", "pt-", "pr-", "pb-", "pl-",
			"m-", "mx-", "my-", "mt-", "mr-", "mb-", "ml-",
			"space-",

			// Typography
			"font-", "text-", "italic", "not-italic", "antialiased", "subpixel-antialiased",
			"tracking-", "leading-", "list-", "color-scheme-", "text-shadow-",
			"uppercase", "lowercase", "capitalize", "normal-case",
			"truncate", "line-clamp-", "text-ellipsis", "text-clip",
			"align-", "whitespace-", "break-", "content-",

			// Backgrounds (v4 bg-linear-to replacing gradient)
			"bg-", "bg-linear-to-", "bg-gradient-to-", "from-", "via-", "to-",

			// Borders & Outlines (v4 outline-hidden replacing outline-none)
			"rounded-", "border", "border-", "divide-", "outline-", "outline-hidden", "outline-none", "ring-",

			// Effects & Filters
			"shadow-", "opacity-", "mix-blend-", "bg-blend-",
			"filter", "blur-", "brightness-", "contrast-", "drop-shadow-", "grayscale-", "hue-rotate-", "invert-", "saturate-", "sepia-",
			"backdrop-",

			// Tables
			"border-collapse", "border-spacing-", "table-layout-", "caption-side-",

			// Transitions & Animations
			"transition", "duration-", "ease-", "delay-", "animate-",

			// Transforms (v4 3D Additions)
			"transform", "scale-", "rotate-", "translate-", "skew-", "origin-",
			"perspective-", "perspective-origin-", "backface-", "transform-style-",

			// Interactivity
			"accent-", "appearance-", "cursor-", "caret-", "pointer-events-", "resize", "scroll-", "snap-", "touch-", "select-", "will-change-",

			// SVG
			"fill-", "stroke-", "stroke-width-",
		},
		VariantOrder: map[string]int{
			// Container Queries & Breakpoints
			"@3xs": 10, "@2xs": 11, "@xs": 12, "@sm": 13, "@md": 14, "@lg": 15, "@xl": 16, "@2xl": 17, "@3xl": 18, "@4xl": 19, "@5xl": 20, "@6xl": 21, "@7xl": 22,
			"@min-": 25, "@max-": 26,
			"sm": 30, "md": 31, "lg": 32, "xl": 33, "2xl": 34,
			"max-sm": 35, "max-md": 36, "max-lg": 37, "max-xl": 38, "max-2xl": 39,

			// Themes & Media
			"dark": 50, "light": 51,
			"print": 52, "screen": 53,
			"portrait": 54, "landscape": 55,
			"motion-safe": 56, "motion-reduce": 57,
			"contrast-more": 58, "contrast-less": 59,
			"forced-colors": 60,

			// Dynamic Prefixes (e.g. group-hover, has-[.foo])
			"group-": 70, "peer-": 71, "has-": 72, "not-": 73, "aria-": 74, "data-": 75, "supports-": 76,

			// Standard States
			"first": 80, "last": 81, "only": 82, "odd": 83, "even": 84,
			"first-of-type": 85, "last-of-type": 86, "only-of-type": 87,
			"visited": 88, "target": 89, "open": 90, "default": 91,
			"checked": 92, "indeterminate": 93, "placeholder-shown": 94,
			"autofill": 95, "optional": 96, "required": 97,
			"valid": 98, "invalid": 99, "in-range": 100, "out-of-range": 101,
			"read-only": 102, "empty": 103,

			// Interaction
			"focus-within": 110, "hover": 111, "focus": 112, "focus-visible": 113, "active": 114,
			"disabled": 115, "enabled": 116, "inert": 117,

			// v4 specifics
			"starting": 120,

			// Pseudo-elements
			"before": 130, "after": 131, "first-letter": 132, "first-line": 133,
			"marker": 134, "selection": 135, "file": 136, "backdrop": 137,
			"placeholder": 138,
		},
		FilePatterns:    []string{".html"},
		ClassAttributes: []string{"class"},
	}
}

func (config *Config) merge(userConfig *UserConfig) {
	if len(userConfig.FilePatterns) > 0 {
		config.FilePatterns = userConfig.FilePatterns
	}

	if len(userConfig.ClassAttributes) > 0 {
		config.ClassAttributes = userConfig.ClassAttributes
	}

	sort.Slice(config.ClassOrder, func(i, j int) bool {
		return len(config.ClassOrder[i]) > len(config.ClassOrder[j])
	})
}
