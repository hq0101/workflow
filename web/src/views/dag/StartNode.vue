<template>
  <div className="editable-sample-start-node">
    <div
      className="editable-sample-start-node-move-hand"
      @mousedown="startNodeDragging"
    />
    <Connector
      type="out"
      shape="square"
      :node="node"
      :niceDag="niceDag"
    />
  </div>
</template>

<script>
import Connector from "./Connector.vue";
export default {
  components: { Connector },
  props: ["niceDagReactive", "node"],
  setup(props) {
    return {
      niceDag: props.niceDagReactive.use(),
      startNodeDragging(e) {
        props.niceDag.startNodeDragging(props.node, e);
      },
    };
  },
};
</script>
