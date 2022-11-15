import { createRouter, createWebHistory } from 'vue-router'
import Home from "@/components/app1/Home";
import About from "@/components/app1/About";
import Contact from "@/components/app1/Contact";
import Login from "@/components/app1/Login";

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home,
  },
  {
    path: '/about',
    name: 'About',
    component: About,
  },
  {
    path: '/contact',
    name: 'Contact',
    component: Contact,
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
  },

]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
