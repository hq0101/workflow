<template>
  <div className="editable-sample-node">
    <GroupControl
      v-if="showGroupCtrl()"
      :node="node"
      :niceDag="niceDagReactive.use()"
    />
    <NodeControl
      v-if="!showGroupCtrl()"
      :node="node"
      :niceDag="niceDagReactive.use()"
    />
    <Connector type="in" />
    <Connector type="out" :node="node" :niceDag="niceDagReactive.use()" />
  </div>
</template>

<script>
import Connector from "./Connector.vue";
import GroupControl from "./GroupControl.vue";
import NodeControl from "./NodeControl.vue";

export default {
  name: "EditableNode",
  props: ["node", "niceDagReactive"],
  components: {
    Connector,
    GroupControl,
    NodeControl,
  },
  setup(props) {
    return {
      showGroupCtrl() {
        return props.node.children?.length > 0 && !props.node.collapse;
      },
    };
  },
};
</script>
