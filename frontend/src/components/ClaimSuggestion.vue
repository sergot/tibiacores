<template>
  <div class="claim-suggestion">
    <h3>Character Name Already Registered</h3>
    <p>The character "{{ characterName }}" is already registered to another user.</p>
    <div class="options">
      <div class="option">
        <h4>Is this your character?</h4>
        <p>
          If this is your character, you can claim ownership by verifying it with your Tibia.com
          character comment.
        </p>
        <router-link
          :to="{ name: 'claim-character', query: { character: characterName } }"
          class="button primary"
        >
          Start Claim Process
        </router-link>
      </div>
      <div class="option">
        <h4>Not your character?</h4>
        <p>If this is not your character, please try with a different character name.</p>
        <button @click="$emit('try-different')" class="button secondary">
          Try Different Character
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'ClaimSuggestion',
  props: {
    characterName: {
      type: String,
      required: true,
    },
  },
  emits: ['try-different'],
})
</script>

<style scoped>
.claim-suggestion {
  max-width: 600px;
  margin: 20px auto;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: #f9f9f9;
}

.options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-top: 20px;
}

.option {
  padding: 15px;
  border: 1px solid #eee;
  border-radius: 4px;
  background: white;
}

.option h4 {
  margin: 0 0 10px 0;
  color: #333;
}

.option p {
  margin: 0 0 15px 0;
  color: #666;
  font-size: 0.9rem;
}

.button {
  display: inline-block;
  padding: 8px 16px;
  border-radius: 4px;
  text-decoration: none;
  text-align: center;
  cursor: pointer;
  border: none;
  width: 100%;
  font-weight: 500;
  transition: opacity 0.2s;
}

.button.primary {
  background-color: #4caf50;
  color: white;
}

.button.secondary {
  background-color: #f5f5f5;
  color: #333;
  border: 1px solid #ddd;
}

.button:hover {
  opacity: 0.9;
}
</style>
