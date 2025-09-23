// Firebase Configuration for Corridor OS
import { initializeApp } from "firebase/app";
import { getAnalytics } from "firebase/analytics";
import { getFirestore } from "firebase/firestore";
import { getFunctions } from "firebase/functions";

// Your web app's Firebase configuration
const firebaseConfig = {
  apiKey: "AIzaSyDH0PdHHHAViB_nAeJ23JFfvd-3bTmnXGE",
  authDomain: "corridor-os.firebaseapp.com",
  projectId: "corridor-os",
  storageBucket: "corridor-os.firebasestorage.app",
  messagingSenderId: "829280519801",
  appId: "1:829280519801:web:73620328fa6235cc8e76a8",
  measurementId: "G-19PRG33J4Y"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const analytics = getAnalytics(app);
const db = getFirestore(app);
const functions = getFunctions(app);

// Export for use in Corridor OS
window.firebaseApp = app;
window.firebaseAnalytics = analytics;
window.firebaseDB = db;
window.firebaseFunctions = functions;

// Track Corridor OS usage
if (typeof gtag !== 'undefined') {
  gtag('event', 'corridor_os_load', {
    event_category: 'engagement',
    event_label: 'Professional OS Launch'
  });
}

console.log('🔥 Firebase initialized for Corridor OS Professional');

