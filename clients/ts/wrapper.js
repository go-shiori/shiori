// Wrapper to expose the API client as a global variable
import * as ShioriAPIModule from './dist/esm/index.js';

// Expose to global scope
window.ShioriAPI = ShioriAPIModule;