<script setup>
import { ref } from 'vue'
import AssetItem from './AssetItem.vue'

defineProps({
  msg: String
})

const name = ref("")

const count = ref(0)

const Items = ref([])

const getData = () => {
  fetch('http://localhost:8090/assets')
    .then((response) => response.json())
    .then((data) => {
      console.log(data)
      return  data.filter(x => x.Record.Owner === name.value)
    })
    .then((data) => Items.value = data);
}
</script>

<template>
  <input v-model="name"/> <button v-on:click="getData()">search</button>
  <hr>
  <!-- {{Items}} -->
  <div>
    <AssetItem v-for="item in Items" 
      :id="item.Key"
      :color="item.Record.Color"
      :size="item.Record.Size"
      :owner="item.Record.Owner"
      :value="item.Record.AppraisedValue"
      />  
  </div>

</template>

<style scoped>
.read-the-docs {
  color: #888;
}
</style>
