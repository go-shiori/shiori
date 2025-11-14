<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue';
import { useI18n } from 'vue-i18n';
import { computed, onMounted, ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useRouter, useRoute } from 'vue-router';
import { AdminIcon } from '@/components/icons';
import { SystemApi, AccountsApi } from '@/client';
import type { ApiV1InfoResponse, ModelAccountDTO } from '@/client';
import { getApiConfig } from '@/utils/api-config';

const { t } = useI18n();
const authStore = useAuthStore();
const router = useRouter();
const route = useRoute();

// Tab management - determine active tab from route
const activeTab = computed(() => {
  if (route.path === '/admin/users') return 'users';
  return 'system';
});

// System information state
const systemInfo = ref<ApiV1InfoResponse | null>(null);
const systemInfoLoading = ref(false);
const systemInfoError = ref<string | null>(null);

// User management state
const accounts = ref<ModelAccountDTO[]>([]);
const accountsLoading = ref(false);
const accountsError = ref<string | null>(null);

// Check if user is admin/owner
const isAdmin = computed(() => !!authStore.user?.owner);

// Load system information
const loadSystemInfo = async () => {
  if (!authStore.token) return;

  systemInfoLoading.value = true;
  systemInfoError.value = null;

  try {
    const api = new SystemApi(getApiConfig(authStore.token));
    systemInfo.value = await api.apiV1SystemInfoGet();
  } catch (error: any) {
    console.error('Failed to load system info:', error);
    systemInfoError.value =
      error.message || 'Failed to load system information';
  } finally {
    systemInfoLoading.value = false;
  }
};

// Load accounts
const loadAccounts = async () => {
  if (!authStore.token) return;

  accountsLoading.value = true;
  accountsError.value = null;

  try {
    const api = new AccountsApi(getApiConfig(authStore.token));
    accounts.value = await api.apiV1AccountsGet();
  } catch (error: any) {
    console.error('Failed to load accounts:', error);
    accountsError.value = error.message || 'Failed to load accounts';
  } finally {
    accountsLoading.value = false;
  }
};

// Redirect to library if not admin
onMounted(async () => {
  if (!isAdmin.value) {
    router.push('/library');
    return;
  }

  // Load system information for admin users
  await loadSystemInfo();
});
</script>

<template>
  <AppLayout v-if="isAdmin">
    <template #header>
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold">{{ t('admin.title') }}</h1>
      </div>
    </template>

    <div class="space-y-6">
      <!-- Tab Navigation -->
      <div class="bg-white dark:bg-gray-800 rounded-md shadow-sm">
        <div class="border-b border-gray-200 dark:border-gray-700">
          <nav class="-mb-px flex space-x-8 px-6" aria-label="Tabs">
            <button
              @click="router.push('/admin')"
              :class="[
                activeTab === 'system'
                  ? 'border-red-500 text-red-600 dark:text-red-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600',
                'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center',
              ]"
            >
              <svg
                class="w-5 h-5 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              {{ t('admin.system_info') }}
            </button>
            <button
              @click="
                router.push('/admin/users');
                loadAccounts();
              "
              :class="[
                activeTab === 'users'
                  ? 'border-red-500 text-red-600 dark:text-red-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600',
                'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center',
              ]"
            >
              <svg
                class="w-5 h-5 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"
                />
              </svg>
              {{ t('admin.user_management') }}
            </button>
          </nav>
        </div>

        <!-- Tab Content -->
        <div class="p-6">
          <!-- System Information Tab -->
          <div v-if="activeTab === 'system'">
            <!-- Loading state -->
            <div
              v-if="systemInfoLoading"
              class="flex items-center text-sm text-blue-600 dark:text-blue-400"
            >
              <svg
                class="animate-spin -ml-1 mr-2 h-4 w-4"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              {{ t('common.loading') }}
            </div>

            <!-- Error state -->
            <div
              v-else-if="systemInfoError"
              class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-md"
            >
              <div class="flex items-center">
                <svg
                  class="w-5 h-5 mr-2"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fill-rule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clip-rule="evenodd"
                  />
                </svg>
                {{ systemInfoError }}
              </div>
            </div>

            <!-- System information content -->
            <div v-else-if="systemInfo" class="space-y-4">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Version Information -->
                <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
                  <h3 class="font-medium text-gray-900 dark:text-white mb-3">
                    {{ t('admin.version_info') }}
                  </h3>
                  <div class="space-y-2 text-sm">
                    <div class="flex justify-between">
                      <span class="text-gray-600 dark:text-gray-400">
                        {{ t('admin.shiori_version') }}:
                      </span>
                      <span class="font-mono text-gray-900 dark:text-white">
                        {{ systemInfo.version?.tag || 'Unknown' }}
                      </span>
                    </div>
                    <div
                      v-if="systemInfo.version?.commit"
                      class="flex justify-between"
                    >
                      <span class="text-gray-600 dark:text-gray-400">
                        {{ t('admin.commit') }}:
                      </span>
                      <span class="font-mono text-gray-900 dark:text-white">
                        {{ systemInfo.version.commit.substring(0, 8) }}
                      </span>
                    </div>
                    <div
                      v-if="systemInfo.version?.date"
                      class="flex justify-between"
                    >
                      <span class="text-gray-600 dark:text-gray-400">
                        {{ t('admin.build_date') }}:
                      </span>
                      <span class="text-gray-900 dark:text-white">
                        {{ systemInfo.version.date }}
                      </span>
                    </div>
                  </div>
                </div>

                <!-- System Details -->
                <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
                  <h3 class="font-medium text-gray-900 dark:text-white mb-3">
                    {{ t('admin.system_details') }}
                  </h3>
                  <div class="space-y-2 text-sm">
                    <div class="flex justify-between">
                      <span class="text-gray-600 dark:text-gray-400">
                        {{ t('admin.database_engine') }}:
                      </span>
                      <span class="text-gray-900 dark:text-white">
                        {{ systemInfo.database || 'Unknown' }}
                      </span>
                    </div>
                    <div class="flex justify-between">
                      <span class="text-gray-600 dark:text-gray-400">
                        {{ t('admin.operating_system') }}:
                      </span>
                      <span class="text-gray-900 dark:text-white">
                        {{ systemInfo.os || 'Unknown' }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- User Management Tab -->
          <div v-else-if="activeTab === 'users'">
            <!-- Loading state -->
            <div
              v-if="accountsLoading"
              class="flex items-center text-sm text-blue-600 dark:text-blue-400"
            >
              <svg
                class="animate-spin -ml-1 mr-2 h-4 w-4"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              {{ t('common.loading') }}
            </div>

            <!-- Error state -->
            <div
              v-else-if="accountsError"
              class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-md"
            >
              <div class="flex items-center">
                <svg
                  class="w-5 h-5 mr-2"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fill-rule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clip-rule="evenodd"
                  />
                </svg>
                {{ accountsError }}
              </div>
            </div>

            <!-- Accounts list -->
            <div v-else-if="accounts.length > 0" class="space-y-4">
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium text-gray-900 dark:text-white">
                  {{ t('admin.accounts') }}
                </h3>
                <button
                  class="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                >
                  {{ t('admin.add_account') }}
                </button>
              </div>

              <div
                class="bg-gray-50 dark:bg-gray-700 rounded-lg overflow-hidden"
              >
                <div class="divide-y divide-gray-200 dark:divide-gray-600">
                  <div
                    v-for="account in accounts"
                    :key="account.id"
                    class="p-4 flex items-center justify-between"
                  >
                    <div class="flex items-center">
                      <div class="flex-shrink-0">
                        <div
                          class="w-8 h-8 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center"
                        >
                          <svg
                            class="w-5 h-5 text-gray-600 dark:text-gray-400"
                            fill="currentColor"
                            viewBox="0 0 20 20"
                          >
                            <path
                              fill-rule="evenodd"
                              d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                              clip-rule="evenodd"
                            />
                          </svg>
                        </div>
                      </div>
                      <div class="ml-3">
                        <p
                          class="text-sm font-medium text-gray-900 dark:text-white"
                        >
                          {{ account.username }}
                          <span
                            v-if="account.owner"
                            class="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
                          >
                            {{ t('admin.owner') }}
                          </span>
                        </p>
                        <p class="text-sm text-gray-500 dark:text-gray-400">
                          ID: {{ account.id }}
                        </p>
                      </div>
                    </div>
                    <div class="flex items-center space-x-2">
                      <button
                        class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                        :title="t('admin.change_password')"
                      >
                        <svg
                          class="w-5 h-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
                          />
                        </svg>
                      </button>
                      <button
                        v-if="!account.owner"
                        class="text-gray-400 hover:text-red-600 dark:hover:text-red-400"
                        :title="t('admin.delete_account')"
                      >
                        <svg
                          class="w-5 h-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- No accounts -->
            <div v-else class="text-center py-8">
              <svg
                class="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"
                />
              </svg>
              <h3
                class="mt-2 text-sm font-medium text-gray-900 dark:text-white"
              >
                {{ t('admin.no_accounts') }}
              </h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.no_accounts_description') }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
