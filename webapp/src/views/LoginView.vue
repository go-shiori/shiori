<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useI18n } from 'vue-i18n';

// Props for destination
const props = defineProps<{
  dst?: string
}>();

const { t } = useI18n();
const username = ref('');
const password = ref('');
const rememberMe = ref(false);
const errorMessage = ref('');
const isLoading = ref(false);
const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

// Check if already authenticated on mount
onMounted(async () => {
  // If we already have a token, validate it
  if (authStore.token) {
    isLoading.value = true;
    const isValid = await authStore.validateToken();
    isLoading.value = false;

    if (isValid) {
      // If valid, redirect to destination or home
      redirectAfterLogin();
    }
  }
});

const login = async () => {
  if (!username.value || !password.value) {
    errorMessage.value = t('auth.login_failed');
    return;
  }

  isLoading.value = true;
  errorMessage.value = '';

  try {
    const success = await authStore.login(username.value, password.value, rememberMe.value);

    if (success) {
      // Redirect to destination or home
      redirectAfterLogin();
    } else {
      // Display the error message from the auth store
      errorMessage.value = authStore.error || t('auth.login_failed');
    }
  } catch (error: any) {
    console.error('Login error:', error);
    errorMessage.value = error.message || t('auth.login_failed');
  } finally {
    isLoading.value = false;
  }
};

// Helper function to redirect after successful login
const redirectAfterLogin = () => {
  // First check the store for a destination
  let destination = authStore.getAndClearRedirectDestination();

  // If no destination in store, check props and route query
  if (!destination) {
    destination = props.dst || route.query.dst as string || '/home';
  }

  // Redirect to the destination
  router.push(destination);
};
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900">
    <div class="w-full max-w-md bg-white dark:bg-gray-800 shadow-lg rounded-md overflow-hidden">
      <!-- Logo and Header -->
      <div class="bg-red-500 text-white py-6 px-4 text-center">
        <div class="text-4xl font-bold mb-1">栞 shiori</div>
        <div class="text-sm">simple bookmark manager</div>
      </div>

      <!-- Login Form -->
      <div class="p-8">
        <div v-if="errorMessage"
          class="mb-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 rounded-md text-sm text-center">
          {{ errorMessage }}
        </div>

        <div v-if="isLoading && authStore.token"
          class="mb-4 p-3 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-md text-sm text-center">
          {{ t('common.loading') }}
        </div>

        <form @submit.prevent="login">
          <div class="mb-6">
            <div class="flex items-center mb-4">
              <div class="w-28 text-right mr-4 text-gray-700 dark:text-gray-300">{{ t('auth.username') }}:</div>
              <input v-model="username" type="text"
                class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500"
                :placeholder="t('auth.username')" required />
            </div>

            <div class="flex items-center">
              <div class="w-28 text-right mr-4 text-gray-700 dark:text-gray-300">{{ t('auth.password') }}:</div>
              <input v-model="password" type="password"
                class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500"
                :placeholder="t('auth.password')" required />
            </div>
          </div>

          <div class="flex justify-center items-center mb-6">
            <input id="remember-me" v-model="rememberMe" type="checkbox"
              class="h-4 w-4 text-red-500 focus:ring-red-500 border-gray-300 dark:border-gray-600 rounded" />
            <label for="remember-me" class="ml-2 block text-sm text-gray-700 dark:text-gray-300">{{
              t('auth.remember_me') }}</label>
          </div>

          <div class="flex justify-center">
            <button type="submit"
              class="w-full bg-gray-800 dark:bg-gray-700 text-white py-2 px-4 rounded-md hover:bg-gray-700 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 uppercase font-medium"
              :disabled="isLoading">
              <span v-if="isLoading">{{ t('common.loading') }}</span>
              <span v-else>{{ t('auth.login') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Additional custom styles if needed */
</style>
