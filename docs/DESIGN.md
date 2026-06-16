# Design Standards: Premium AAA Aesthetics

This document defines the visual and UX language for the AITriage Dashboard. We aim for a "High-Utility, High-Hype" interface that impresses at first glance while remaining technically rigorous.

## 1. Color Palette (Cyber-Obsidian)
We avoid flat, generic colors.
- **Background**: `#0a0a0c` (Obsidian Black)
- **Primary**: `#4f46e5` (Electric Indigo)
- **Secondary**: `#9333ea` (Cyber Purple)
- **Critical**: `#ef4444` (Vibrant Crimson)
- **Highlight**: `linear-gradient(90deg, #4f46e5, #9333ea, #ec4899)` (The "Dash" Gradient)

## 2. Typography
- **Primary**: `Outfit` (or `Inter`) - Selected for its modern, clean, but characterful letterforms.
- **Mono**: `JetBrains Mono` - For code snippets and technical IDs.

## 3. UI Patterns
- **Glassmorphism**: Cards should have a translucent background (`rgba(23, 23, 26, 0.7)`) with a subtle `1px` border (`rgba(255, 255, 255, 0.05)`) and `12px` backdrop blur.
- **Micro-Animations**: 
    - Hover effects on cards (border-color transition + subtle scale).
    - Pulse animations for "Critical" status markers.
- **Spacing**: Generous padding (`p-8` for containers) to allow the data "to breathe."

## 4. Tone of Voice
- **Findings**: Be direct, technical, and urgent.
- **Suggestions**: Be actionable and authoritative.
- **Agentic reasoning**: Prefix with `[Agentic Reasoning]` to emphasize the "Brain" involvement.

## 5. Information Hierarchy
1. **The Security Grade**: The first thing the user sees. Large, centered, or top-right.
2. **The Pulse (Metrics)**: Summary numbers (Critical count, Entropy level).
3. **The Feed**: Detailed findings sorted by severity.
4. **The Remedy**: Direct links to automated fixes.
