package config

import (
	"fmt"
	"os"

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
		ClassOrder: []string{"box-border", "box-content", "block", "inline-block", "inline", "flex", "inline-flex", "table", "inline-table",
			"table-caption", "table-cell", "table-column", "table-column-group", "table-footer-group", "table-header-group",
			"table-row-group", "table-row", "flow-root", "grid", "inline-grid", "contents", "list-item", "hidden", "float-",
			"clear-", "isolate", "isolation-auto", "object-", "overflow-", "overscroll-", "static", "fixed", "absolute",
			"relative", "sticky", "top-", "right-", "bottom-", "left-", "inset-", "visible", "invisible", "z-",
			"flex-basis-", "flex-direction-", "flex-wrap-", "flex-", "flex-grow", "flex-shrink", "order-", "grid-cols-",
			"grid-col-", "grid-rows-", "grid-row-", "grid-flow-", "gap-", "justify-", "justify-items-", "justify-self-",
			"items-", "align-", "place-content-", "place-items-", "place-self-",
			"p-", "px-", "py-", "pt-", "pr-", "pb-", "pl-", "m-", "mx-", "my-", "mt-", "mr-", "mb-", "ml-", "space-",
			"w-", "min-w-", "max-w-", "h-", "min-h-", "max-h-",
			"font-", "text-", "italic", "not-italic", "font-weight-", "font-variant-numeric-", "letter-spacing-",
			"line-clamp-", "line-height-", "list-", "text-align-", "text-color-", "text-decoration-",
			"text-decoration-color-", "text-decoration-style-", "text-decoration-thickness-", "text-underline-offset-",
			"text-transform-", "text-overflow-", "text-indent-", "vertical-align-", "whitespace-", "break-",
			"content-",
			"bg-", "bg-opacity-", "bg-origin-", "bg-position-", "bg-repeat-", "bg-size-", "bg-image-", "gradient-to-",
			"from-", "via-", "to-",
			"rounded-", "border", "border-", "border-opacity-", "border-style-", "divide-", "divide-opacity-",
			"divide-style-", "outline-", "outline-offset-", "outline-style-", "ring-", "ring-offset-", "ring-opacity-",
			"shadow-", "opacity-", "mix-blend-", "bg-blend-",
			"filter", "blur-", "brightness-", "contrast-", "drop-shadow-", "grayscale-", "hue-rotate-", "invert-",
			"saturate-", "sepia-", "backdrop-",
			"border-collapse", "border-spacing-", "table-layout-", "caption-side-",
			"transition", "duration-", "ease-", "delay-", "animate-",
			"transform", "scale-", "rotate-", "translate-", "skew-", "transform-origin-",
			"accent-", "appearance-", "cursor-", "caret-", "pointer-events-", "resize", "scroll-", "scroll-snap-",
			"touch-", "select-", "will-change-",
			"fill-", "stroke-", "stroke-width-",
			"sr-only", "not-sr-only"},
		VariantOrder: map[string]int{"sm": 0, "md": 1, "lg": 2, "xl": 3, "2xl": 4, "dark": 10,
			"motion-safe": 20, "motion-reduce": 21, "portrait": 22, "landscape": 23,
			"first": 30, "last": 31, "odd": 32, "even": 33, "visited": 34, "checked": 35,
			"disabled": 36, "enabled": 37, "hover": 40, "focus": 41, "focus-within": 42,
			"focus-visible": 43, "active": 44},
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
}
