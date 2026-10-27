<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { Bucket, Node } from '../lib/models'
echarts.use([LineChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])
const props = defineProps<{ buckets: Bucket[]; width: number; nodes: Node[] }>()
const host = ref<HTMLElement>()
let chart: echarts.ECharts | undefined
let observer: ResizeObserver
function render() {
  if (!chart) return
  const ids = [...new Set(props.buckets.map((r) => r.node_id))]
  const series = ids.map((id) => {
    const values = props.buckets.filter((r) => r.node_id === id).sort((a, b) => a.at - b.at)
    const data: [number, number | null][] = []
    values.forEach((r, i) => {
      // 插入空点，禁止在缺测窗口上绘制连续曲线。
      const previous = values[i - 1]
      if (previous && r.at - previous.at > props.width)
        data.push([(previous.at + props.width) * 1000, null])
      data.push([r.at * 1000, r.success > 0 ? r.latency_ms : null])
    })
    return {
      name: props.nodes.find((n) => n.id === id)?.name || id,
      type: 'line',
      showSymbol: values.length < 10,
      connectNulls: false,
      data,
    }
  })
  chart.setOption(
    {
      color: ['#087f73', '#6090b8', '#b28747', '#997aab'],
      tooltip: { trigger: 'axis' },
      legend: { bottom: 0 },
      grid: { left: 55, right: 20, top: 20, bottom: 55 },
      xAxis: { type: 'time', axisLine: { lineStyle: { color: '#dce4e2' } } },
      yAxis: { type: 'value', name: 'ms', splitLine: { lineStyle: { color: '#edf1ee' } } },
      series,
    },
    true,
  )
}
onMounted(() => {
  if (host.value) {
    chart = echarts.init(host.value)
    observer = new ResizeObserver(() => chart?.resize())
    observer.observe(host.value)
    render()
  }
})
watch(() => props.buckets, render)
onUnmounted(() => {
  observer?.disconnect()
  chart?.dispose()
})
</script>
<template>
  <div
    ref="host"
    style="height: 280px; width: 100%"
    role="img"
    aria-label="Latency history by node"
  />
</template>
