<template>
  <nav class="navbar navbar-expand-lg bg-light">
    <div class="container-fluid">
      <router-link class="navbar-brand" to="/">GO UA</router-link>
      <div class="collapse navbar-collapse" id="navbarNavDropdown">
        <ul class="navbar-nav">
          <li class="nav-item">
            <router-link class="nav-link" to="/">Home</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/about">About</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/contact">Contact us </router-link>
          </li>
        </ul>
      </div>
      <router-link to="/login" class="nav-link" v-if="store.token == ''">Login</router-link>
      <a style="color:gray; text-decoration: none;" href="javascript:void(0)" v-else @click="logout()">Logout</a>
    </div>
  </nav>
</template>

<script>
import {store} from "@/components/store";
import Security from "@/components/Security";
import router from "@/router";

export default {
  name: "HeaderComponent",
  data() {
    return {
      store,
    }
  },
  methods: {
    logout() {
      const payload = { token: store.token }

      fetch(process.env.VUE_APP_API_URL + "/users/logout", Security.requestOptions(payload))
          .then((response) => response.json())
          .then((response) => {
            if (response.error == true) {
              console.log(response.error)
            } else {
              store.token = ""
              store.user = {}
              document.cookie =
                  '_site_data=; Path=/; ' +
                  'SameSite=Strict; Secure; ' +
                  'Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
              router.push("/login")
            }
          })

    }
  }
}
</script>
