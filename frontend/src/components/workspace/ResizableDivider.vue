<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps<{
  position: number // percentage (0-100)
  minPosition?: number
  maxPosition?: number
  direction?: 'horizontal' | 'vertical'
}>()

const emit = defineEmits<{
  (e: 'update:position', value: number): void
  (e: 'drag-start'): void
  (e: 'drag-end'): void
}>()

const isDragging = ref(false)
const dividerRef = ref<HTMLElement | null>(null)

const minPos = props.minPosition ?? 10
const maxPos = props.maxPosition ?? 90
const isHorizontal = props.direction !== 'vertical'

function startDrag(event: MouseEvent) {
  event.preventDefault()
  isDragging.value = true
  emit('drag-start')
  
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
  document.body.style.cursor = isHorizontal ? 'row-resize' : 'col-resize'
  document.body.style.userSelect = 'none'
}

function onDrag(event: MouseEvent) {
  if (!isDragging.value || !dividerRef.value) return
  
  const parent = dividerRef.value.parentElement
  if (!parent) return
  
  const rect = parent.getBoundingClientRect()
  let newPosition: number
  
  if (isHorizontal) {
    // Horizontal divider (splits top/bottom)
    const offsetY = event.clientY - rect.top
    newPosition = (offsetY / rect.height) * 100
  } else {
    // Vertical divider (splits left/right)
    const offsetX = event.clientX - rect.left
    newPosition = (offsetX / rect.width) * 100
  }
  
  // Clamp position
  newPosition = Math.max(minPos, Math.min(maxPos, newPosition))
  emit('update:position', newPosition)
}

function stopDrag() {
  isDragging.value = false
  emit('drag-end')
  
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

// Handle touch events for mobile
function startTouchDrag(event: TouchEvent) {
  event.preventDefault()
  isDragging.value = true
  emit('drag-start')
  
  document.addEventListener('touchmove', onTouchDrag, { passive: false })
  document.addEventListener('touchend', stopTouchDrag)
}

function onTouchDrag(event: TouchEvent) {
  if (!isDragging.value || !dividerRef.value) return
  event.preventDefault()
  
  const touch = event.touches[0]
  if (!touch) return
  
  const parent = dividerRef.value.parentElement
  if (!parent) return
  
  const rect = parent.getBoundingClientRect()
  let newPosition: number
  
  if (isHorizontal) {
    const offsetY = touch.clientY - rect.top
    newPosition = (offsetY / rect.height) * 100
  } else {
    const offsetX = touch.clientX - rect.left
    newPosition = (offsetX / rect.width) * 100
  }
  
  newPosition = Math.max(minPos, Math.min(maxPos, newPosition))
  emit('update:position', newPosition)
}

function stopTouchDrag() {
  isDragging.value = false
  emit('drag-end')
  
  document.removeEventListener('touchmove', onTouchDrag)
  document.removeEventListener('touchend', stopTouchDrag)
}

// Cleanup on unmount
onUnmounted(() => {
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  document.removeEventListener('touchmove', onTouchDrag)
  document.removeEventListener('touchend', stopTouchDrag)
})
</script>

<template>
  <div
    ref="dividerRef"
    class="resizable-divider"
    :class="{ 
      dragging: isDragging,
      horizontal: isHorizontal,
      vertical: !isHorizontal
    }"
    @mousedown="startDrag"
    @touchstart="startTouchDrag"
  >
    <div class="divider-handle">
      <div class="handle-dots">
        <span></span>
        <span></span>
        <span></span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.resizable-divider {
  position: relative;
  flex-shrink: 0;
  background: #e4e7ed;
  transition: background 0.2s;
  z-index: 10;
}

.resizable-divider.horizontal {
  height: 6px;
  cursor: row-resize;
}

.resizable-divider.vertical {
  width: 6px;
  cursor: col-resize;
}

.resizable-divider:hover,
.resizable-divider.dragging {
  background: #409eff;
}

.divider-handle {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
}

.horizontal .divider-handle {
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 100%;
}

.vertical .divider-handle {
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 100%;
  height: 40px;
}

.handle-dots {
  display: flex;
  gap: 3px;
}

.horizontal .handle-dots {
  flex-direction: row;
}

.vertical .handle-dots {
  flex-direction: column;
}

.handle-dots span {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: #909399;
  transition: background 0.2s;
}

.resizable-divider:hover .handle-dots span,
.resizable-divider.dragging .handle-dots span {
  background: #fff;
}
</style>
