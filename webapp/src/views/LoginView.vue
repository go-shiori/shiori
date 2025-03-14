<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const username = ref('');
const password = ref('');
const rememberMe = ref(false);
const router = useRouter();

const login = async () => {
  // TODO: Implement actual login logic with API
  console.log('Login with:', { username: username.value, password: password.value, rememberMe: rememberMe.value });

  // Store a dummy token for testing
  localStorage.setItem('token', 'dummy-token');

  // Redirect to home
  router.push('/home');
};
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="w-full max-w-md bg-white shadow-lg rounded-md overflow-hidden">
      <!-- Logo and Header -->
      <div class="bg-red-500 text-white py-6 px-4 text-center">
        <div class="text-4xl font-bold mb-1">栞 shiori</div>
        <div class="text-sm">simple bookmark manager</div>
      </div>

      <!-- Login Form -->
      <div class="p-8">
        <div class="mb-6">
          <div class="flex items-center mb-4">
            <div class="w-28 text-right mr-4">Username:</div>
            <input v-model="username" type="text"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500"
              placeholder="Username" required />
          </div>

          <div class="flex items-center">
            <div class="w-28 text-right mr-4">Password:</div>
            <input v-model="password" type="password"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500"
              placeholder="Password" required />
          </div>
        </div>

        <div class="flex justify-center items-center mb-6">
          <input id="remember-me" v-model="rememberMe" type="checkbox"
            class="h-4 w-4 text-red-500 focus:ring-red-500 border-gray-300 rounded" />
          <label for="remember-me" class="ml-2 block text-sm text-gray-700">Remember me</label>
        </div>

        <div class="flex justify-center">
          <button @click="login" type="button"
            class="w-full bg-gray-800 text-white py-2 px-4 rounded-md hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 uppercase font-medium">
            LOG IN
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Additional custom styles if needed */
</style>
