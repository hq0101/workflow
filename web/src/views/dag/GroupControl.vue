<template>
  <div className="editable-sample-node-group-content" @mouseenter="onMouseEnter" @mouseleave="onMouseLeaveDelay">
    <div>
      <div
        className="editable-sample-node-group-content-bar"
        @mousedown="startNodeDragging"
      >
        <span>{{ node.data?.label || node.id }}</span>
        <SvgIcon v-if="niceDag.editing"><MoveSvg /></SvgIcon>
      </div>
      <div v-show="isExpandButtons">
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
      <MyButton @click="shrinkNode">
        <ShrinkSvg />
      </MyButton>
    </div>
  </div>
</template>

<script>
import SvgIcon from "./svgs/SvgIcon.vue";
import MyButton from "./svgs/MyButton.vue";
import MoveSvg from "./svgs/move.vue";
import ShrinkSvg from "./svgs/shrink.vue";
import { ref } from 'vue'

export default {
  name: "EditableGroupControl",
  components: { SvgIcon, MoveSvg, MyButton, ShrinkSvg },
  props: ["node", "niceDag", "type", "shape"],
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
      shrinkNode() {
        props.node.shrink();
      },
      removeNode() {
        props.node.remove();
      },
      isExpandButtons,
    };
  },
};
</script>
