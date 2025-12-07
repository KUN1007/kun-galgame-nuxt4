<script setup lang="ts"></script>

<template>
  <KunCard
    :is-transparent="false"
    :is-hoverable="false"
    :is-pressable="false"
    color="success"
  >
    <h3 class="text-success-800 text-base sm:text-lg">鲲论坛的使用小提示</h3>
    <div class="mt-3 space-y-3 text-sm">
      <p>1. 网站有些内容默认不会显示, 请在右上角设置按钮打开 NSFW 模式</p>
      <p>
        2. 如果您想要寻找 Galgame 补丁资源, 请前往
        <KunLink size="sm" target="_blank" :to="kungal.domain.patch">
          鲲 Galgame 补丁 (moyu.moe)
        </KunLink>
        , 本论坛是 Galgame 论坛社区网站, Galgame
        补丁资源需要在我们的补丁站下载。
      </p>
      <p>
        3. 欢迎加入我们的
        <KunLink
          :to="kungal.domain.telegram_group"
          size="sm"
          target="_blank"
          rel="noopener noreferrer"
        >
          Telegram 群组
        </KunLink>
        进行关于小只可爱软萌女孩子的交流
      </p>
      <p>
        4. 鲲论坛代码全部自己开发, 如果您觉得我们网站帮助到了您, 欢迎给我们的
        <KunLink
          :to="kungal.github"
          size="sm"
          target="_blank"
          rel="noopener noreferrer"
        >
          GitHub 仓库
        </KunLink>
        点一个 star 以支持我们的开源工作!
      </p>
    </div>
  </KunCard>
</template>
