#!/usr/bin/env python3
"""
Autonomous SVG Star History Generator for Cybermes.
Fetches exact stargazer timestamps via GitHub API (using GITHUB_TOKEN or gh CLI)
and generates a responsive SVG chart into assets/star_history.svg.
"""

import os
import sys
import json
import urllib.request
import subprocess
from datetime import datetime

REPO = os.environ.get('GITHUB_REPOSITORY', 'Zyrexnn/Cybermes')
TOKEN = os.environ.get('GITHUB_TOKEN') or os.environ.get('GH_TOKEN')

if not TOKEN:
    try:
        TOKEN = subprocess.check_output(['gh', 'auth', 'token']).decode('utf-8').strip()
    except Exception:
        TOKEN = None

def fetch_stargazers():
    stars = []
    page = 1
    headers = {
        'Accept': 'application/vnd.github.v3.star+json',
        'User-Agent': 'Cybermes-Star-Chart-Bot'
    }
    if TOKEN:
        headers['Authorization'] = f'Bearer {TOKEN}'

    print(f'[+] Fetching stargazers for {REPO}...')
    while True:
        url = f'https://api.github.com/repos/{REPO}/stargazers?per_page=100&page={page}'
        req = urllib.request.Request(url, headers=headers)
        try:
            with urllib.request.urlopen(req) as resp:
                data = json.loads(resp.read().decode('utf-8'))
                if not data:
                    break
                stars.extend(data)
                if len(data) < 100:
                    break
                page += 1
        except Exception as e:
            print(f'[-] Error on page {page}: {e}')
            break

    if not stars:
        print('[!] No stars retrieved from API.')
        return []

    print(f'[+] Successfully fetched {len(stars)} total stars across {page} page(s).')
    
    date_counts = {}
    for item in stars:
        d = item['starred_at'][:10]
        date_counts[d] = date_counts.get(d, 0) + 1

    sorted_dates = sorted(date_counts.keys())
    data_points = []
    total = 0
    for d in sorted_dates:
        total += date_counts[d]
        data_points.append((d, total))
    return data_points

def generate_svg(data_points, output_path='assets/star_history.svg'):
    if not data_points:
        return

    width = 850
    height = 360
    padding_left = 65
    padding_right = 45
    padding_top = 55
    padding_bottom = 50

    plot_width = width - padding_left - padding_right
    plot_height = height - padding_top - padding_bottom

    max_stars = max([c for _, c in data_points] + [250])
    y_max = ((max_stars + 49) // 50) * 50

    n = len(data_points)
    points = []
    for i, (date_str, count) in enumerate(data_points):
        x = padding_left + (i / (n - 1)) * plot_width if n > 1 else padding_left + plot_width / 2
        y = padding_top + plot_height - (count / y_max) * plot_height
        dt = datetime.strptime(date_str, '%Y-%m-%d')
        formatted_date = dt.strftime('%d %b')
        points.append((x, y, count, formatted_date))

    line_d = f'M {points[0][0]:.1f} {points[0][1]:.1f}'
    area_d = f'M {points[0][0]:.1f} {padding_top + plot_height:.1f} L {points[0][0]:.1f} {points[0][1]:.1f}'

    for i in range(1, len(points)):
        x0, y0 = points[i-1][0], points[i-1][1]
        x1, y1 = points[i][0], points[i][1]
        cx1 = x0 + (x1 - x0) / 2
        cy1 = y0
        cx2 = x0 + (x1 - x0) / 2
        cy2 = y1
        line_d += f' C {cx1:.1f} {cy1:.1f}, {cx2:.1f} {cy2:.1f}, {x1:.1f} {y1:.1f}'
        area_d += f' C {cx1:.1f} {cy1:.1f}, {cx2:.1f} {cy2:.1f}, {x1:.1f} {y1:.1f}'

    area_d += f' L {points[-1][0]:.1f} {padding_top + plot_height:.1f} Z'

    y_ticks = 5
    grid_lines = []
    for i in range(y_ticks + 1):
        val = int(i * (y_max / y_ticks))
        y_pos = padding_top + plot_height - (val / y_max) * plot_height
        grid_lines.append(f'<line x1="{padding_left}" y1="{y_pos:.1f}" x2="{width - padding_right}" y2="{y_pos:.1f}" stroke="#30363d" stroke-dasharray="3,3" stroke-width="1" /><text x="{padding_left - 12}" y="{y_pos + 4:.1f}" fill="#8b949e" font-family="-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif" font-size="12" text-anchor="end">{val}</text>')

    x_labels = []
    circles = []
    badges = []
    for x, y, count, formatted_date in points:
        x_labels.append(f'<text x="{x:.1f}" y="{height - padding_bottom + 25}" fill="#8b949e" font-family="-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif" font-size="12" text-anchor="middle">{formatted_date}</text>')
        circles.append(f'<circle cx="{x:.1f}" cy="{y:.1f}" r="5" fill="#58a6ff" stroke="#0d1117" stroke-width="2" /><circle cx="{x:.1f}" cy="{y:.1f}" r="9" fill="#58a6ff" fill-opacity="0.3" />')
        badges.append(f'<rect x="{x - 26:.1f}" y="{y - 30:.1f}" width="52" height="22" rx="6" fill="#161b22" stroke="#58a6ff" stroke-width="1" /><text x="{x:.1f}" y="{y - 15:.1f}" fill="#58a6ff" font-family="-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif" font-size="11" font-weight="bold" text-anchor="middle">★ {count}</text>')

    latest_count = data_points[-1][1]

    svg = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="100%" height="100%">
  <defs>
    <linearGradient id="gradientArea" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#58a6ff" stop-opacity="0.45" />
      <stop offset="70%" stop-color="#1f6feb" stop-opacity="0.15" />
      <stop offset="100%" stop-color="#1f6feb" stop-opacity="0.0" />
    </linearGradient>
    <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" stop-color="#79c0ff" />
      <stop offset="100%" stop-color="#58a6ff" />
    </linearGradient>
    <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
      <feGaussianBlur stdDeviation="3" result="blur" />
      <feMerge>
        <feMergeNode in="blur" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    </filter>
  </defs>

  <rect width="{width}" height="{height}" rx="12" fill="#0d1117" stroke="#30363d" stroke-width="1.5" />

  <text x="{padding_left}" y="32" fill="#f0f6fc" font-family="-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif" font-size="16" font-weight="bold">⭐ GitHub Star Growth Timeline</text>
  <text x="{width - padding_right}" y="32" fill="#7ee787" font-family="-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif" font-size="14" font-weight="bold" text-anchor="end">🚀 Total: {latest_count} Stars</text>

  {''.join(grid_lines)}

  <path d="{area_d}" fill="url(#gradientArea)" />
  <path d="{line_d}" fill="none" stroke="url(#lineGrad)" stroke-width="3" filter="url(#glow)" />

  {''.join(x_labels)}
  {''.join(circles)}
  {''.join(badges)}
</svg>"""

    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(svg)
    print(f'[+] SVG chart successfully written to: {output_path}')

if __name__ == '__main__':
    pts = fetch_stargazers()
    if pts:
        generate_svg(pts)
