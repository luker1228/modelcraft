#!/usr/bin/env python3
"""
Icon Slicer — 将 AI 生成的图标图片切分为单独的透明 PNG 文件
用法:
  python3 slice_icons.py --input sprite.png --output ./icons/ --rows 4 --cols 5 --names icon1,icon2,...
"""

import argparse
import os
import sys
from pathlib import Path


def remove_white_background(img, threshold=240):
    """将白色/近白色背景转为透明"""
    img = img.convert("RGBA")
    data = img.getdata()
    new_data = []
    for r, g, b, a in data:
        if r >= threshold and g >= threshold and b >= threshold:
            new_data.append((r, g, b, 0))
        else:
            new_data.append((r, g, b, a))
    img.putdata(new_data)
    return img


def slice_icons(
    input_path: str,
    output_dir: str,
    rows: int,
    cols: int,
    names: list[str] | None = None,
    padding: int = 4,
    size: int = 24,
    remove_bg: bool = True,
):
    try:
        from PIL import Image
    except ImportError:
        print("错误: 需要安装 Pillow。运行: pip3 install Pillow")
        sys.exit(1)

    img = Image.open(input_path)
    img_w, img_h = img.size
    print(f"图片尺寸: {img_w}x{img_h}")

    cell_w = img_w // cols
    cell_h = img_h // rows
    print(f"每格尺寸: {cell_w}x{cell_h} (padding={padding})")

    os.makedirs(output_dir, exist_ok=True)

    idx = 0
    saved = []
    for row in range(rows):
        for col in range(cols):
            left = col * cell_w + padding
            upper = row * cell_h + padding
            right = (col + 1) * cell_w - padding
            lower = (row + 1) * cell_h - padding

            cell = img.crop((left, upper, right, lower))

            if remove_bg:
                cell = remove_white_background(cell)
            else:
                cell = cell.convert("RGBA")

            if size and size != (right - left):
                cell = cell.resize((size, size), Image.LANCZOS)

            # 确定文件名
            if names and idx < len(names):
                name = names[idx].strip()
            else:
                name = f"icon-{row:02d}-{col:02d}"

            # 转换名称格式: CamelCase -> kebab-case
            if not name.startswith("icon-"):
                # 将 CamelCase 转为 kebab-case
                import re
                kebab = re.sub(r"(?<!^)(?=[A-Z])", "-", name).lower()
                filename = f"icon-{kebab}.png"
            else:
                filename = f"{name}.png"

            out_path = os.path.join(output_dir, filename)
            cell.save(out_path, "PNG")
            saved.append(filename)
            print(f"  ✓ {filename}")
            idx += 1

    print(f"\n完成！共切分 {len(saved)} 个图标 → {output_dir}")
    return saved


def main():
    parser = argparse.ArgumentParser(
        description="将 AI 生成的图标拼图切分为单独的透明 PNG"
    )
    parser.add_argument("--input", "-i", required=True, help="输入图片路径")
    parser.add_argument(
        "--output",
        "-o",
        default="modelcraft-front/public/icons/",
        help="输出目录 (默认: modelcraft-front/public/icons/)",
    )
    parser.add_argument("--rows", "-r", type=int, required=True, help="图标网格行数")
    parser.add_argument("--cols", "-c", type=int, required=True, help="图标网格列数")
    parser.add_argument(
        "--names",
        "-n",
        help="图标名称列表，逗号分隔，按行优先顺序 (例: Sparkles,Database,Shield)",
    )
    parser.add_argument(
        "--padding",
        "-p",
        type=int,
        default=4,
        help="每格内边距像素 (默认: 4，用于去除间隔线)",
    )
    parser.add_argument(
        "--size",
        "-s",
        type=int,
        default=24,
        help="输出图标尺寸 px (默认: 24，可选 32/48/64)",
    )
    parser.add_argument(
        "--no-remove-bg",
        action="store_true",
        help="不去除白色背景 (默认会将白色转透明)",
    )

    args = parser.parse_args()

    names = None
    if args.names:
        names = args.names.split(",")
        expected = args.rows * args.cols
        if len(names) != expected:
            print(
                f"警告: 提供了 {len(names)} 个名称，但网格共 {expected} 格 ({args.rows}x{args.cols})"
            )

    slice_icons(
        input_path=args.input,
        output_dir=args.output,
        rows=args.rows,
        cols=args.cols,
        names=names,
        padding=args.padding,
        size=args.size,
        remove_bg=not args.no_remove_bg,
    )


if __name__ == "__main__":
    main()
