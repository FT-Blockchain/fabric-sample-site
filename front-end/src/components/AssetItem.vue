<script setup>
  import { $vfm, VueFinalModal, ModalsContainer } from 'vue-final-modal'
  import { ref } from 'vue'

  const props = defineProps({
    id: String,
    color: String,
    size: String,
    owner: String,
    value: String
  })

    const showModal = ref(false)
    const name = ref("")

    const getData = () => {
      fetch('http://localhost:8090/transaction', {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          asset_id: props.id,
          owner: name.value
        })
      });
      
      showModal.value = false
    }
</script>
    
    <template>
      <div class="asset-Item">
        ID: {{props.id}}
        colour: {{props.color}}
        size: {{props.size}}
        owner: {{props.owner}}
        value: {{props.value}}

        <button @click="showModal = true">Transact</button>
      </div>

      <vue-final-modal v-model="showModal">
        <div class="modal-content">
          Who do you want to give this asset?
          <div>
            <input v-model="name"/> <button v-on:click="getData()">Send</button>
            <button v-on:click="showModal = false">Close</button>
          </div>
          
        </div>
        
      </vue-final-modal>

    </template>
    
<style scoped>
  .asset-Item {
    background-color: #888;
  }

  ::v-deep .modal-content {
    display: inline-block;
    padding: 1rem;
    border: 1px solid #e2e8f0;
    border-radius: 0.25rem;
    background: #fff;
    margin-top: 15%;
    height: 250px;
    width: 250px;
  }
  .modal__title {
    font-size: 1.5rem;
    font-weight: 700;
  }
</style>
    