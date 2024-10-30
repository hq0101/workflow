<template>
  <div className="editable-sample-node-content" @mouseenter="onMouseEnter" @mouseleave="onMouseLeaveDelay">
    <div
      className="editable-sample-node-content-bar"
      @mousedown="startNodeDragging"
    >
      <span>{{ node.data?.label || node.id }}</span>
      <SvgIcon v-if="niceDag.editing"><MoveSvg /></SvgIcon>
    </div>
    <div v-show="isExpandButtons">
      <div
        className="editable-sample-node-content-setting-button"
        role="button"
      >
        <el-tooltip
          class="box-item"
          effect="dark"
          content="配置节点"
          placement="top"
        >
          <el-icon><Setting /></el-icon>
        </el-tooltip>
      </div>
      <div
        className="editable-sample-node-content-copy-button"
        role="button"
      >
        <el-tooltip
          class="box-item"
          effect="dark"
          content="复制节点"
          placement="top"
        >
          <el-icon><CopyDocument /></el-icon>
        </el-tooltip>
      </div>
      <div
        className="editable-sample-node-content-delete-button"
        role="button"
      >
        <el-tooltip
          class="box-item"
          effect="dark"
          content="删除节点"
          placement="top"
        >
          <el-popconfirm
            title="确定要删除吗？"
            placement="top"
            @confirm="removeNode"
          >
            <template #reference>
              <el-icon><Delete /></el-icon>
            </template>
          </el-popconfirm>
        </el-tooltip>
      </div>
    </div>
    <MyButton @click="expandNode" v-if="isExpandButtonVisible()">
      <PlusSvg />
    </MyButton>
  </div>
</template>

<script>
import SvgIcon from "./svgs/SvgIcon.vue";
import MyButton from "./svgs/MyButton.vue";
import MoveSvg from "./svgs/move.vue";
import PlusSvg from "./svgs/plus.vue";
import { ref } from 'vue'
import { Delete, CopyDocument, Setting } from '@element-plus/icons-vue'

export default {
  name: "EditableNodeControl",
  components: { Delete, CopyDocument, Setting, SvgIcon, MoveSvg, PlusSvg, MyButton },
  props: ["node", "niceDag"],
  setup(props) {
    const isExpandButtons = ref(false)
    let hideTimer = null;
    return {
      startNodeDragging(e) {
        props.niceDag.startNodeDragging(props.node, e);
      },
      onMouseEnter() {
        if (hideTimer) {
          clearTimeout(hideTimer);
          hideTimer = null;
        }
        isExpandButtons.value = true;
      },
      onMouseLeaveDelay() {
        hideTimer = setTimeout(() => {
          isExpandButtons.value = false;
        }, 500);
      },
      removeNode() {
        props.node.remove();
      },
      expandNode() {
        props.node.expand();
      },
      isExpandButtonVisible() {
        return (
          props.node.children?.length > 0 ||
          props.node.data?.lazyLoadingChildren
        );
      },
      isExpandButtons,
    };
  },
};
</script>
