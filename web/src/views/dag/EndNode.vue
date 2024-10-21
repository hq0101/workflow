<template>
  <div className="editable-sample-end-node" @mousedown="startNodeDragging">
    <Connector type="in" shape="square" />
  </div>
</template>

<script>
import Connector from "./Connector.vue";
export default {
  components: { Connector },
  props: ["niceDag", "node"],
  setup(props) {
    return {
      startNodeDragging(e) {
        props.niceDag.startNodeDragging(props.node, e);
      },
    };
  },
};
</script>
