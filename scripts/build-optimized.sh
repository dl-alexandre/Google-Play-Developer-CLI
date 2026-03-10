#!/bin/bash
# Build optimization script for gpd - Google Play Developer CLI
# Target: Reduce binary size from ~30MB to ~15-20MB
#
# Usage: ./scripts/build-optimized.sh [level]
#   level: basic | optimized | compressed (default: optimized)

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARY_DIR="$PROJECT_DIR/bin"
BINARY_NAME="gpd"
BUILD_LOG="$BINARY_DIR/build-sizes.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Level can be: basic, optimized, compressed
LEVEL="${1:-optimized}"

# Version information
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

# Print header
print_header() {
    echo ""
    echo "========================================"
    echo "  gpd Binary Size Optimization"
    echo "  Target: ~15-20MB (from ~30MB)"
    echo "========================================"
    echo ""
}

# Get human-readable size
get_size() {
    local file="$1"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        stat -f%z "$file" 2>/dev/null || echo "0"
    else
        # Linux
        stat -c%s "$file" 2>/dev/null || echo "0"
    fi
}

# Format size to human-readable
format_size() {
    local size=$1
    if command -v numfmt >/dev/null 2>&1; then
        numfmt --to=iec-i --suffix=B "$size"
    else
        if [ "$size" -lt 1024 ]; then
            echo "${size}B"
        elif [ "$size" -lt 1048576 ]; then
            echo "$((size / 1024))KB"
        else
            echo "$((size / 1048576))MB"
        fi
    fi
}

# Print size comparison
print_comparison() {
    local name="$1"
    local size="$2"
    local baseline="$3"
    local reduction=$((baseline - size))
    local pct=$((100 * reduction / baseline))
    
    echo "  $name: $(format_size "$size") (${pct}% reduction)"
}

# Test binary functionality
test_binary() {
    local binary="$1"
    local name="$2"
    
    echo ""
    echo -e "${BLUE}Testing $name binary...${NC}"
    
    if [ ! -f "$binary" ]; then
        echo -e "${RED}✗ Binary not found: $binary${NC}"
        return 1
    fi
    
    # Test version command
    if "$binary" version >/dev/null 2>&1; then
        echo -e "${GREEN}✓ $name binary works (version command successful)${NC}"
        return 0
    else
        echo -e "${RED}✗ $name binary failed (version command)${NC}"
        return 1
    fi
}

# Clean previous builds
clean_builds() {
    echo "Cleaning previous builds..."
    rm -rf "$BINARY_DIR"
    mkdir -p "$BINARY_DIR"
}

# Build: Standard (baseline)
build_standard() {
    echo ""
    echo -e "${YELLOW}[1/4] Building STANDARD version (baseline)...${NC}"
    
    local output="$BINARY_DIR/${BINARY_NAME}-standard"
    local ldflags="-X github.com/dl-alexandre/gpd/pkg/version.Version=$VERSION \
        -X github.com/dl-alexandre/gpd/pkg/version.GitCommit=$GIT_COMMIT \
        -X github.com/dl-alexandre/gpd/pkg/version.BuildTime=$BUILD_TIME"
    
    go build -ldflags "$ldflags" -o "$output" ./cmd/gpd
    
    STANDARD_SIZE=$(get_size "$output")
    echo "  Standard size: $(format_size "$STANDARD_SIZE")"
    
    test_binary "$output" "standard"
}

# Build: Basic optimization (-s -w)
build_basic() {
    echo ""
    echo -e "${YELLOW}[2/4] Building BASIC optimized version (-s -w)...${NC}"
    
    local output="$BINARY_DIR/${BINARY_NAME}-basic"
    local ldflags="-s -w \
        -X github.com/dl-alexandre/gpd/pkg/version.Version=$VERSION \
        -X github.com/dl-alexandre/gpd/pkg/version.GitCommit=$GIT_COMMIT \
        -X github.com/dl-alexandre/gpd/pkg/version.BuildTime=$BUILD_TIME"
    
    go build -ldflags "$ldflags" -o "$output" ./cmd/gpd
    
    BASIC_SIZE=$(get_size "$output")
    print_comparison "Basic" "$BASIC_SIZE" "$STANDARD_SIZE"
    
    test_binary "$output" "basic"
}

# Build: Optimized (-s -w -trimpath, CGO_ENABLED=0)
build_optimized() {
    echo ""
    echo -e "${YELLOW}[3/4] Building FULLY OPTIMIZED version (-s -w -trimpath, CGO_ENABLED=0)...${NC}"
    
    local output="$BINARY_DIR/${BINARY_NAME}-optimized"
    local ldflags="-s -w \
        -X github.com/dl-alexandre/gpd/pkg/version.Version=$VERSION \
        -X github.com/dl-alexandre/gpd/pkg/version.GitCommit=$GIT_COMMIT \
        -X github.com/dl-alexandre/gpd/pkg/version.BuildTime=$BUILD_TIME"
    
    CGO_ENABLED=0 go build -trimpath -ldflags "$ldflags" -o "$output" ./cmd/gpd
    
    OPTIMIZED_SIZE=$(get_size "$output")
    print_comparison "Optimized" "$OPTIMIZED_SIZE" "$STANDARD_SIZE"
    
    # Copy to main binary location
    cp "$output" "$BINARY_DIR/$BINARY_NAME"
    
    test_binary "$output" "optimized"
}

# Build: Compressed with UPX (if available)
build_compressed() {
    echo ""
    echo -e "${YELLOW}[4/4] Building COMPRESSED version (upx)...${NC}"
    
    if ! command -v upx >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠ upx not installed. Skipping compression step.${NC}"
        echo "  To install: brew install upx (macOS) or apt-get install upx (Linux)"
        COMPRESSED_SIZE=$OPTIMIZED_SIZE
        return 0
    fi
    
    local input="$BINARY_DIR/${BINARY_NAME}-optimized"
    local output="$BINARY_DIR/${BINARY_NAME}-compressed"
    
    # Copy first, then compress
    cp "$input" "$output"
    
    # Try different UPX levels
    echo "  Trying UPX compression..."
    if upx --best --lzma -q "$output" 2>/dev/null; then
        COMPRESSED_SIZE=$(get_size "$output")
        print_comparison "Compressed" "$COMPRESSED_SIZE" "$STANDARD_SIZE"
        test_binary "$output" "compressed"
    else
        echo -e "${YELLOW}⚠ UPX compression failed (possibly not supported on this binary)${NC}"
        COMPRESSED_SIZE=$OPTIMIZED_SIZE
    fi
}

# Print summary
print_summary() {
    echo ""
    echo "========================================"
    echo -e "${GREEN}  BUILD OPTIMIZATION SUMMARY${NC}"
    echo "========================================"
    echo ""
    printf "  %-20s %10s %10s\n" "Build Type" "Size" "Reduction"
    printf "  %-20s %10s %10s\n" "----------" "----" "---------"
    printf "  %-20s %10s %10s\n" "Standard (baseline)" "$(format_size $STANDARD_SIZE)" "-"
    
    local basic_reduction=$((100 * (STANDARD_SIZE - BASIC_SIZE) / STANDARD_SIZE))
    printf "  %-20s %10s %10s\n" "Basic (-s -w)" "$(format_size $BASIC_SIZE)" "${basic_reduction}%"
    
    local opt_reduction=$((100 * (STANDARD_SIZE - OPTIMIZED_SIZE) / STANDARD_SIZE))
    printf "  %-20s %10s %10s\n" "Optimized" "$(format_size $OPTIMIZED_SIZE)" "${opt_reduction}%"
    
    if [ "$COMPRESSED_SIZE" -lt "$OPTIMIZED_SIZE" ]; then
        local comp_reduction=$((100 * (STANDARD_SIZE - COMPRESSED_SIZE) / STANDARD_SIZE))
        printf "  %-20s %10s %10s\n" "UPX Compressed" "$(format_size $COMPRESSED_SIZE)" "${comp_reduction}%"
    fi
    
    echo ""
    echo "  Recommended: Use 'optimized' build for distribution"
    echo "  Binary location: $BINARY_DIR/$BINARY_NAME"
    echo ""
    
    # Log results
    {
        echo "Build completed at $(date)"
        echo "Version: $VERSION ($GIT_COMMIT)"
        echo "Standard:  $(format_size $STANDARD_SIZE)"
        echo "Basic:     $(format_size $BASIC_SIZE) (${basic_reduction}% reduction)"
        echo "Optimized: $(format_size $OPTIMIZED_SIZE) (${opt_reduction}% reduction)"
        if [ "$COMPRESSED_SIZE" -lt "$OPTIMIZED_SIZE" ]; then
            local comp_pct=$((100 * (STANDARD_SIZE - COMPRESSED_SIZE) / STANDARD_SIZE))
            echo "Compressed: $(format_size $COMPRESSED_SIZE) (${comp_pct}% reduction)"
        fi
    } > "$BUILD_LOG"
}

# Main execution
main() {
    print_header
    
    # Check if we're in the right directory
    if [ ! -f "$PROJECT_DIR/go.mod" ]; then
        echo -e "${RED}Error: go.mod not found. Run from project root.${NC}"
        exit 1
    fi
    
    cd "$PROJECT_DIR"
    
    # Clean previous builds
    clean_builds
    
    # Build based on level
    case "$LEVEL" in
        standard)
            build_standard
            ;;
        basic)
            build_standard
            build_basic
            ;;
        optimized|all)
            build_standard
            build_basic
            build_optimized
            build_compressed
            print_summary
            ;;
        compressed)
            build_standard
            build_basic
            build_optimized
            build_compressed
            print_summary
            ;;
        *)
            echo -e "${RED}Error: Unknown level '$LEVEL'${NC}"
            echo "Usage: $0 [standard|basic|optimized|compressed|all]"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}✓ Build optimization complete!${NC}"
}

# Run main
main
