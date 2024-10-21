<template>
  <div className="editable-sample-node-content">
    <div
      className="editable-sample-node-content-bar"
      @mousedown="startNodeDragging"
    >
      <span>{{ node.data?.label || node.id }}</span>
      <SvgIcon v-if="niceDag.editing"><MoveSvg /></SvgIcon>
    </div>
    <div
      className="editable-sample-node-content-delete-button"
      role="button"
      @mousedown="removeNode"
    >
      <SvgIcon><CloseSvg /></SvgIcon>
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
import CloseSvg from "./svgs/close.vue";
import PlusSvg from "./svgs/plus.vue";

export default {
  name: "EditableNodeControl",
  components: { SvgIcon, MoveSvg, CloseSvg, PlusSvg, MyButton },
  props: ["node", "niceDag"],
  setup(props) {
    return {
      startNodeDragging(e) {
        props.niceDag.startNodeDragging(props.node, e);
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
    };
  },
};
</script>
