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
		ClassOrder: []string{
			// daisyUI Skeleton
			"skeleton",

			// daisyUI Button
			"btn", "btn-primary", "btn-secondary", "btn-accent", "btn-neutral", "btn-info", "btn-success", "btn-warning",
			"btn-error", "btn-outline", "btn-dash", "btn-soft", "btn-ghost", "btn-link", "btn-active", "btn-disabled", "btn-xs",
			"btn-sm", "btn-md", "btn-lg", "btn-xl", "btn-wide", "btn-block", "btn-square", "btn-circle",

			// daisyUI Dropdown
			"dropdown", "dropdown-content", "dropdown-start", "dropdown-center", "dropdown-end", "dropdown-top", "dropdown-bottom",
			"dropdown-left", "dropdown-right", "dropdown-hover", "dropdown-open",

			// daisyUI Fab / Speed Dial
			"fab", "fab-close", "fap-main-action", "fab-flower",

			// daisyUI Modal
			"modal", "modal-box", "modal-action", "modal-backdrop", "modal-toggle", "modal-open", "modal-top", "modal-middle",
			"modal-bottom", "mdoal-start", "modal-end",

			// daisyUI Swap
			"swap", "swap-on", "swap-off", "swap-indeterminate", "swap-active", "swap-rotate", "swap-flip",

			// daisyUI Accordion / Collapse
			"collapse", "collapse-title", "collapse-content", "collapse-arrow", "collapse-plus", "collapse-open", "collapse-close",

			// daisyUI Avatar
			"avatar", "avatar-group", "avatar-online", "avatar-offline", "avatar-placeholder",

			// daisyUI Badge
			"badge", "badge-outline", "badge-dash", "badge-soft", "badge-ghost", "badge-primary", "badge-secondary",
			"badge-accent", "badge-neutral", "badge-info", "badge-success", "badge-warning", "badge-error", "badge-xs",
			"badge-sm", "badge-md", "badge-lg", "badge-xl",

			// daisyUI Card
			"card", "card-title", "card-body", "card-actions", "card-border", "card-dash", "card-side", "image-full", "card-xs",
			"card-sm", "card-md", "card-lg", "card-xl",

			// daisyUI Carousel
			"carousel", "carousel-item", "carousel-start", "carousel-center", "carousel-end", "carousel-horizontal", "carousel-vertical",

			// daisyUI Chat Bubble
			"chat", "chat-image", "chat-header", "chat-footer", "chat-bubble", "chat-start", "chat-end",
			"chat-bubble-primary", "chat-bubble-secondary", "chat-bubble-accent", "chat-bubble-neutral", "chat-bubble-info",
			"chat-bubble-success", "chat-bubble-warning", "chat-bubble-error",

			// daisyUI Countdown
			"countdown",

			// daisyUI Diff
			"diff", "diff-item-1", "diff-item-2", "diff-resizer",

			// daisyUI Hover Gallery
			"hover-gallery",

			// daisyUI KBD
			"kbd", "kbd-xs", "kbd-sm", "kbd-md", "kbd-lg", "kbd-xl",

			// daisyUI List
			"list", "list-row", "list-col-wrap", "list-col-grow",

			// daisyUI Stat
			"stats", "stat", "stat-title", "stat-value", "stat-desc", "stat-figure", "stat-actions", "stats-horizontal",
			"stats-vertical",

			// daisyUI Status
			"status", "status-primary", "status-secondary", "status-accent", "status-neutral", "status-info", "status-success",
			"status-warning", "status-error", "status-xs", "status-sm", "status-md", "status-lg", "status-xl",

			// daisyUI Table
			"table", "table-zebra", "table-pin-rows", "table-pin-cols", "table-xs", "table-sm", "table-md", "table-lg", "table-xl",

			// daisyUI Timeline
			"timeline", "timeline-start", "timeline-middle", "timeline-end", "timeline-snap-icon", "timeline-box",
			"timeline-compact", "timeline-horizontal", "timeline-vertical",

			// daisyUI Breadcrumbs
			"breadcrumbs",

			// daisyUI Dock
			"dock", "dock-label", "dock-active", "dock-xs", "dock-sm", "dock-md", "dock-lg", "dock-xl",

			// daisyUI Link
			"link", "link-hover", "link-primary", "link-secondary", "link-accent", "link-neutral", "link-success",
			"link-info", "link-warning", "link-error",

			// daisyUI Menu
			"menu", "menu-title", "menu-dropdown", "menu-dropdown-toggle", "menu-disabled", "menu-active", "menu-focus",
			"menu-dropdown-show", "menu-xs", "menu-sm", "menu-md", "menu-lg", "menu-xl", "menu-horizontal", "menu-vertical",

			// daisyUI Navbar
			"navbar", "navbar-start", "navbar-center", "navbar-end",

			// daisyUI Pagination / Join
			"join", "join-item", "join-horizontal", "join-vertical",

			// daisyUI Steps
			"steps", "step", "step-icon", "step-primary", "step-secondary", "step-accent", "step-neutral", "step-info",
			"step-success", "step-warning", "step-error", "step-horizontal", "step-vertical",

			// daisyUI Tabs
			"tabs", "tab", "tab-content", "tabs-box", "tabs-border", "tabs-lift", "tab-active", "tab-disabled", "tabs-top",
			"tabs-bottom", "tabs-xs", "tabs-sm", "tabs-md", "tabs-lg", "tabs-xl",

			// daisyUI Alert
			"alert", "alert-outline", "alert-dash", "alert-soft", "alert-ghost", "alert-info", "alert-success",
			"alert-warning", "alert-error", "alert-horizontal", "alert-vertical",

			// daisyUI Loading
			"loading", "loading-spinner", "loading-dots", "loading-ring", "loading-ball", "loading-bars", "loading-infinity",
			"loading-xs", "loading-sm", "loading-md", "loading-lg", "loading-xl",

			// daisyUI Progress
			"progress", "progress-primary", "progress-secondary", "progress-accent", "progress-neutral", "progress-info",
			"progress-success", "progress-warning", "progress-error",

			// daisyUI Radial Progress
			"radial-progress",

			// daisyUI Toast
			"toast", "toast-start", "toast-center", "toast-end", "toast-top", "toast-middle", "toast-bottom",

			// daisyUI Tooltip
			"tooltip", "tooltip-content", "tooltip-top", "tooltip-bottom", "tooltip-left", "tooltip-right", "tooltip-open",
			"tooltip-primary", "tooltip-secondary", "tooltip-accent", "tooltip-neutral", "tooltip-info", "tooltip-success",
			"tooltip-warning", "tooltip-error",

			// daisyUI Calendar
			"cally", "pika-single", "react-day-picker",

			// daisyUI Checkbox
			"checkbox", "checkbox-primary", "checkbox-secondary", "checkbox-accent", "checkbox-neutral", "checkbox-info",
			"checkbox-success", "checkbox-warning", "checkbox-error", "checkbox-xs", "checkbox-sm", "checkbox-md", "checkbox-lg",
			"checkbox-xl",

			// daisyUI Fieldset
			"fieldset", "fieldset-legend",

			// daisyUI File Input
			"file-input", "file-input-ghost", "file-input-primary", "file-input-secondary", "file-input-accent",
			"file-input-neutral", "file-input-info", "file-input-success", "file-input-warning", "file-input-error",
			"file-input-xs", "file-input-sm", "file-input-md", "file-input-lg", "file-input-xl",

			// daisyUI Field Filter
			"filter", "filter-reset",

			// daisyUI Label
			"label", "floating-label",

			// daisyUI Radio
			"radio", "radio-primary", "radio-secondary", "radio-accent", "radio-neutral", "radio-info", "radio-success",
			"radio-warning", "radio-error", "radio-xs", "radio-sm", "radio-md", "radio-lg", "radio-xl",

			// daisyUI Range Slider
			"range", "range-primary", "range-secondary", "range-accent", "range-neutral", "range-info", "range-success",
			"range-warning", "range-error", "range-xs", "range-sm", "range-md", "range-lg", "range-xl",

			// daisyUI Rating
			"rating", "rating-half", "rating-hidden", "rating-xs", "rating-sm", "rating-md", "rating-lg", "rating-xl",

			// daisyUI Select
			"select", "select-ghost", "select-primary", "select-secondary", "select-accent", "select-neutral", "select-info",
			"select-success", "select-warning", "select-error", "select-xs", "select-sm", "select-md", "select-lg", "select-xl",

			// daisyUI Text Input
			"input", "input-ghost", "input-primary", "input-secondary", "input-accent", "input-neutral", "input-info",
			"input-success", "input-warning", "input-error", "input-xs", "input-sm", "input-md", "input-lg", "input-xl",

			// daisyUI Textarea
			"textarea", "textarea-ghost", "textarea-primary", "textarea-secondary", "textarea-accent", "textarea-neutral",
			"textarea-info", "textarea-success", "textarea-warning", "textarea-error", "textarea-xs", "textarea-sm",
			"textarea-md", "textarea-lg", "textarea-xl",

			// daisyUI Toggle
			"toggle", "toggle-primary", "toggle-secondary", "toggle-accent", "toggle-neutral", "toggle-info", "toggle-success",
			"toggle-warning", "toggle-error", "toggle-xs", "toggle-sm", "toggle-md", "toggle-lg", "toggle-xl",

			// daisyUI Validator
			"validator", "validator-hint",

			// daisyUI Divider
			"divider", "divider-primary", "divider-secondary", "divider-accent", "divider-neutral", "divider-info",
			"divider-success", "divider-warning", "divider-error", "divider-start", "divider-end", "divider-horizontal",
			"divider-vertical",

			// daisyUI Drawer
			"drawer", "drawer-toggle", "drawer-content", "drawer-side", "drawer-overlay", "drawer-end", "drawer-open",
			"is-drawer-open:", "is-drawer-close:",

			// daisyUI Footer
			"footer", "footer-title", "footer-center", "footer-horizontal", "footer-vertical",

			// daisyUI Hero
			"hero", "hero-content", "hero-overlay",

			// daisyUI Indicator
			"indicator", "indicator-item", "indicator-start", "indicator-center", "indicator-end", "indicator-top",
			"indicator-middle", "indicator-bottom",

			// daisyUI Mask
			"mask", "mask-squircle", "mask-heart", "mask-hexagon", "mask-hexagon-2", "mask-decagon", "mask-pentagon",
			"mask-diamond", "mask-square", "mask-circle", "mask-star", "mask-star-2", "mask-triangle", "mask-triangle-2",
			"mask-triangle-3", "mask-triangle-4", "mask-half-1", "mask-half-2",

			// daisyUI Stack
			"stack", "stack-top", "stack-bottom", "stack-start", "stack-end",

			// daisyUI Browser
			"mockup-browser", "mockup-browser-toolbar",

			// daisyUI Code
			"mockup-code",

			// daisyUI Phone
			"mockup-phone", "mockup-phone-camera", "mockup-phone-display",

			// daisyUI Window
			"mockup-window",

			// daisyUI Background
			"bg-primary", "bg-secondary", "bg-accent", "bg-neutral", "bg-info", "bg-success", "bg-warning", "bg-error",

			// daisyUI Color Utility
			"to-primary", "to-secondary", "to-accent", "to-neutral", "to-info", "to-success", "to-warning", "to-error",
			"via-primary", "via-secondary", "via-accent", "via-neutral", "via-info", "via-success", "via-warning", "via-error",
			"from-primary", "from-secondary", "from-accent", "from-neutral", "from-info", "from-success", "from-warning", "from-error",
			"ring-primary", "ring-secondary", "ring-accent", "ring-neutral", "ring-info", "ring-success", "ring-warning", "ring-error",
			"fill-primary", "fill-secondary", "fill-accent", "fill-neutral", "fill-info", "fill-success", "fill-warning", "fill-error",
			"caret-primary", "caret-secondary", "caret-accent", "caret-neutral", "caret-info", "caret-success", "caret-warning", "caret-error",
			"stroke-primary", "stroke-secondary", "stroke-accent", "stroke-neutral", "stroke-info", "stroke-success", "stroke-warning", "stroke-error",
			"border-primary", "border-secondary", "border-accent", "border-neutral", "border-info", "border-success", "border-warning", "border-error",
			"divide-primary", "divide-secondary", "divide-accent", "divide-neutral", "divide-info", "divide-success", "divide-warning", "divide-error",
			"accent-primary", "accent-secondary", "accent-accent", "accent-neutral", "accent-info", "accent-success", "accent-warning", "accent-error",
			"shadow-primary", "shadow-secondary", "shadow-accent", "shadow-neutral", "shadow-info", "shadow-success", "shadow-warning", "shadow-error",
			"outline-primary", "outline-secondary", "outline-accent", "outline-neutral", "outline-info", "outline-success", "outline-warning", "outline-error",
			"decoration-primary", "decoration-secondary", "decoration-accent", "decoration-neutral", "decoration-info", "decoration-success", "decoration-warning", "decoration-error",
			"placeholder-primary", "placeholder-secondary", "placeholder-accent", "placeholder-neutral", "placeholder-info", "placeholder-success", "placeholder-warning", "placeholder-error",
			"ring-offset-primary", "ring-offset-secondary", "ring-offset-accent", "ring-offset-neutral", "ring-offset-info", "ring-offset-success", "ring-offset-warning", "ring-offset-error",
			"rounded-box", "rounded-field", "rounded-selector",
			"glass",

			// daisyUI Foreground
			"text-primary", "text-secondary", "text-accent", "text-neutral", "text-info", "text-success", "text-warning", "text-error",

			// daisyUI Theme Controller
			"theme-controller",

			// Layout (Box Sizing, Display, Floats, Clear, Isolation, Object Fit/Position, Overflow, Overscroll, Position, Visibility, Z-Index)
			"box-border", "box-content", "block", "inline-block", "inline", "flex", "inline-flex", "table", "inline-table",
			"table-caption", "table-cell", "table-column", "table-column-group", "table-footer-group", "table-header-group",
			"table-row-group", "table-row", "flow-root", "grid", "inline-grid", "contents", "list-item", "hidden", "float-",
			"clear-", "isolate", "isolation-auto", "object-", "overflow-", "overscroll-", "static", "fixed", "absolute",
			"relative", "sticky", "top-", "right-", "bottom-", "left-", "inset-", "visible", "invisible", "z-",

			// Flexbox & Grid
			"flex-basis-", "flex-direction-", "flex-wrap-", "flex-", "flex-grow", "flex-shrink", "order-", "grid-cols-",
			"grid-col-", "grid-rows-", "grid-row-", "grid-flow-", "gap-", "justify-", "justify-items-", "justify-self-",
			"items-", "align-", "place-content-", "place-items-", "place-self-",

			// Spacing (Padding, Margin, Space Between)
			"p-", "px-", "py-", "pt-", "pr-", "pb-", "pl-", "m-", "mx-", "my-", "mt-", "mr-", "mb-", "ml-", "space-",

			// Sizing (Width, Min-Width, Max-Width, Height, Min-Height, Max-Height)
			"w-", "min-w-", "max-w-", "h-", "min-h-", "max-h-",

			// Typography
			"font-", "text-", "italic", "not-italic", "font-weight-", "font-variant-numeric-", "letter-spacing-",
			"line-clamp-", "line-height-", "list-", "text-align-", "text-color-", "text-decoration-",
			"text-decoration-color-", "text-decoration-style-", "text-decoration-thickness-", "text-underline-offset-",
			"text-transform-", "text-overflow-", "text-indent-", "vertical-align-", "whitespace-", "break-",
			"content-",

			// Backgrounds
			"bg-", "bg-opacity-", "bg-origin-", "bg-position-", "bg-repeat-", "bg-size-", "bg-image-", "gradient-to-",
			"from-", "via-", "to-",

			// Borders
			"rounded-", "border", "border-", "border-opacity-", "border-style-", "divide-", "divide-opacity-",
			"divide-style-", "outline-", "outline-offset-", "outline-style-", "ring-", "ring-offset-", "ring-opacity-",

			// Effects (Box Shadow, Opacity, Mix Blend, Background Blend)
			"shadow-", "opacity-", "mix-blend-", "bg-blend-",

			// Filters (Blur, Brightness, Contrast, Drop Shadow, Grayscale, Hue Rotate, Invert, Saturate, Sepia, Backdrop)
			"filter", "blur-", "brightness-", "contrast-", "drop-shadow-", "grayscale-", "hue-rotate-", "invert-",
			"saturate-", "sepia-", "backdrop-",

			// Tables
			"border-collapse", "border-spacing-", "table-layout-", "caption-side-",

			// Transitions & Animation
			"transition", "duration-", "ease-", "delay-", "animate-",

			// Transforms
			"transform", "scale-", "rotate-", "translate-", "skew-", "transform-origin-",

			// Interactivity
			"accent-", "appearance-", "cursor-", "caret-", "pointer-events-", "resize", "scroll-", "scroll-snap-",
			"touch-", "select-", "will-change-",

			// SVG
			"fill-", "stroke-", "stroke-width-",

			// Screen Readers
			"sr-only", "not-sr-only",
		},
		VariantOrder: map[string]int{
			"sm": 0, "md": 1, "lg": 2, "xl": 3, "2xl": 4, "dark": 10,
			"motion-safe": 20, "motion-reduce": 21, "portrait": 22, "landscape": 23,
			"first": 30, "last": 31, "odd": 32, "even": 33, "visited": 34, "checked": 35,
			"disabled": 36, "enabled": 37, "hover": 40, "focus": 41, "focus-within": 42,
			"focus-visible": 43, "active": 44,
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
}
