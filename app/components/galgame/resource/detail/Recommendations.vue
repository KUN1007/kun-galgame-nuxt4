<script setup lang="ts">
const props = defineProps<{
  recommendations: GalgameResourceCard[]
}>()
</script>

<template>
  <KunCard
    :is-hoverable="false"
    :is-transparent="false"
    content-class="space-y-4 h-full"
  >
    <div v-if="props.recommendations.length" class="flex flex-col gap-3">
      <GalgameResourceCard
        v-for="resource in props.recommendations"
        :key="resource.id"
        :resource="resource"
      />
    </div>
  </KunCard>
</template>
