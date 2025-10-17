set -e

detect_arch() {
  local ARCH="$(uname -m)"

  case $ARCH in
  x86_64 | amd64)
    echo "x64"
    ;;
  aarch64 | arm64)
    echo "arm64"
    ;;
  *)
    echo "Error: Unsupported platform architecture: $ARCH" >&2
    exit 1
    ;;
  esac
}

detect_os() {
  local OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

  case $OS in
  linux)
    echo "linux"
    ;;
  darwin)
    echo "darwin"
    ;;
  mingw | msys | cygwin)
    echo "windows"
    ;;
  *)
    echo "Error: Unsupported OS: $OS" >&2
    exit 1
    ;;
  esac
}

if [ -z "$GOARCH" ]; then
  echo "Warning: No GOARCH provided."
  echo "Resolving..."
  GOARCH="$(detect_arch)"
fi

if [ -z "$GOOS" ]; then
  echo "Warning: No GOOS provided."
  echo "Resolving..."
  GOOS="$(detect_os)"
fi

if [ -z "$VERSION" ]; then
  echo "Warning: No VERSION provided."
  echo "Resolving..."
  VERSION="dev"
fi

BINARY_NAME="tailwind-sorter-$VERSION-$GOOS-$GOARCH"

if [ "$GOOS" == "windows" ]; then
  BINARY_NAME+=".exe"
fi

echo "Building for $VERSION-$GOOS-$GOARCH..."
go build -trimpath -ldflags="-s -w -X 'github.com/selene466/go-tailwind-sorter/cmd.Version=${VERSION}'" -o "dist/${BINARY_NAME}" .
echo "✅ Build complete."

echo "$BINARY_NAME"
