'use client'

import { useEffect, useRef } from 'react'

interface Node {
  x: number
  y: number
  vx: number
  vy: number
  r: number
  pulse: number
  pulseSpeed: number
}

export function NeuralCanvas() {
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    let animId: number
    let nodes: Node[] = []
    const NODE_COUNT = 55
    const MAX_DIST = 130
    const ACCENT = { r: 99, g: 88, b: 238 } // #6358EE — brand purple

    const resize = () => {
      canvas.width = canvas.offsetWidth
      canvas.height = canvas.offsetHeight
    }

    const init = () => {
      resize()
      nodes = Array.from({ length: NODE_COUNT }, () => ({
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        vx: (Math.random() - 0.5) * 0.35,
        vy: (Math.random() - 0.5) * 0.35,
        r: Math.random() * 1.6 + 0.8,
        pulse: Math.random() * Math.PI * 2,
        pulseSpeed: 0.01 + Math.random() * 0.015,
      }))
    }

    const draw = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height)

      // Update positions
      for (const n of nodes) {
        n.x += n.vx
        n.y += n.vy
        n.pulse += n.pulseSpeed
        if (n.x < 0 || n.x > canvas.width) n.vx *= -1
        if (n.y < 0 || n.y > canvas.height) n.vy *= -1
      }

      // Draw edges
      for (let i = 0; i < nodes.length; i++) {
        for (let j = i + 1; j < nodes.length; j++) {
          const a = nodes[i]
          const b = nodes[j]
          const dx = a.x - b.x
          const dy = a.y - b.y
          const dist = Math.sqrt(dx * dx + dy * dy)
          if (dist > MAX_DIST) continue
          const alpha = (1 - dist / MAX_DIST) * 0.22
          ctx.beginPath()
          ctx.moveTo(a.x, a.y)
          ctx.lineTo(b.x, b.y)
          ctx.strokeStyle = `rgba(${ACCENT.r},${ACCENT.g},${ACCENT.b},${alpha})`
          ctx.lineWidth = 0.6
          ctx.stroke()
        }
      }

      // Draw nodes
      for (const n of nodes) {
        const glow = 0.5 + 0.5 * Math.sin(n.pulse)
        const radius = n.r * (1 + 0.3 * glow)

        // Outer glow
        const grad = ctx.createRadialGradient(n.x, n.y, 0, n.x, n.y, radius * 5)
        grad.addColorStop(0, `rgba(${ACCENT.r},${ACCENT.g},${ACCENT.b},${0.25 * glow})`)
        grad.addColorStop(1, `rgba(${ACCENT.r},${ACCENT.g},${ACCENT.b},0)`)
        ctx.beginPath()
        ctx.arc(n.x, n.y, radius * 5, 0, Math.PI * 2)
        ctx.fillStyle = grad
        ctx.fill()

        // Core dot
        ctx.beginPath()
        ctx.arc(n.x, n.y, radius, 0, Math.PI * 2)
        ctx.fillStyle = `rgba(${ACCENT.r},${ACCENT.g},${ACCENT.b},${0.6 + 0.4 * glow})`
        ctx.fill()
      }

      animId = requestAnimationFrame(draw)
    }

    const ro = new ResizeObserver(() => {
      resize()
    })
    ro.observe(canvas)

    init()
    draw()

    return () => {
      cancelAnimationFrame(animId)
      ro.disconnect()
    }
  }, [])

  return (
    <canvas
      ref={canvasRef}
      className="pointer-events-none absolute inset-0 size-full"
      aria-hidden="true"
    />
  )
}
