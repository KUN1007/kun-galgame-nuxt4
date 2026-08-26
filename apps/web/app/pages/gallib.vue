<script setup lang="ts">
const description = `Galgame 信息资料数据库, 收录全部 Galgame 作品资料, 包含本站暂无下载资源的作品, 可按热度, 发售日期, 资料更新排序, 并按发售年份筛选`

useKunSeoMeta({
  title: 'Galgame 信息资料数据库',
  description
})
</script>

<template>
  <GalgameLibraryContainer />
</template>
