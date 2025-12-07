<script setup lang="ts">
import type { KunLayoutCollapsedItem } from '~/constants/layout'

const route = useRoute()

const collapsedItems: Record<string, KunLayoutCollapsedItem[]>[] = [
  {
    create: [
      {
        name: 'create',
        icon: 'lucide:pencil',
        label: '发布主题',
        router: '/edit/topic'
      },
      {
        name: 'createGalgame',
        icon: 'lucide:gamepad-2',
        router: '/edit/galgame/create',
        label: '发布 Gal'
      },
      {
        name: 'createToolset',
        icon: 'lucide:wrench',
        router: '/edit/toolset/create',
        label: '发布 Gal 工具'
      }
    ],
    others: [
      {
        name: 'timeline',
        icon: 'lucide:activity',
        router: '/activity',
        label: '动态时间线'
      },
      {
        name: 'galgame',
        icon: 'lucide:align-end-horizontal',
        router: '/ranking/galgame',
        label: 'Galgame 排行'
      },
      {
        name: 'doc',
        icon: 'lucide:info',
        router: '/doc',
        label: '关于我们'
      },
      {
        name: 'friends',
        icon: 'lucide:handshake',
        router: '/friend-links',
        label: '友情链接'
      }
    ]
  }
]
</script>

<template>
  <div class="flex items-center gap-4">
    <div class="flex w-full gap-4">
      <template v-for="(group, groupIndex) in collapsedItems" :key="groupIndex">
        <template v-for="(items, groupName) in group" :key="groupName">
          <div class="bg-default-500/10 flex items-center gap-2 rounded-lg">
            <KunTooltip
              v-for="item in items"
              :key="item.name"
              :text="item.label"
              position="bottom"
            >
              <KunButton
                is-icon-only
                variant="light"
                size="lg"
                :href="item.router"
                :aria-label="item.label"
                color="default"
                :class="
                  cn(
                    'flex items-center justify-center',
                    route.fullPath === item.router
                      ? 'bg-accent text-primary'
                      : 'text-foreground'
                  )
                "
              >
                <KunIcon :name="item.icon" class="text-lg" />
              </KunButton>
            </KunTooltip>
          </div>
        </template>
      </template>
    </div>
  </div>
</template>
