<template>
  <div style="display: flex; justify-content: center; padding: 200px 0 200px 0;">
    <form>
      <div class="mb-3">
        <p></p>
        <TextInput
            v-model="email"
            label="Email:"
            inputClass="form-control"
            type="email"
            name="email"
            required="true"
        />
      </div>
      <div class="mb-3">
        <TextInput
            v-model="password"
            label="Password:"
            inputClass="form-control"
            type="password"
            name="password"
            required="true"
        />
      </div>
      <input type="" class="btn btn-primary" @click="submitHandler" value="Submit">
    </form>
  </div>
</template>

<script>
import TextInput from "@/components/forms/TextInput";
import Security from "@/components/Security";
import notie from 'notie'
import {store} from "@/components/store";
// import router from "@/router";

export default {
  name: "LoginBody",
  components: {TextInput},
  data() {
    return {
      email: "",
      password: "",
      store,
    }
  },
  methods: {
    submitHandler() {
      const payload = {
        email: this.email,
        password: this.password,
      }
      console.log(payload)
      fetch(process.env.VUE_APP_API_URL + "/users/login", Security.requestOptions(payload))
          .then((response) => response.json())
          .then((response) => {
            if (response.error) {
              notie.alert({
                type: "error",
                text: response.message
              })
            } else {
              store.token = response.data.token.token

              store.user = {
                id: response.data.user.id,
                first_name: response.data.user.first_name,
                last_name: response.data.user.last_name,
                email: response.data.user.email,
              }

              // save info to cookie
              let date = new Date()
              let expDays = 1
              date.setTime(date.getTime() + (expDays) * 24 * 60 * 60 * 1000)

              const expires = "expires" + date.toUTCString()

              // set the cookie
              document.cookie = "_site_data="
                  + JSON.stringify(response.data)
                  + "; "
                  + expires
                  + "; path=/; SameSite=Strict; Secure;"
            }
          })
    }
  }
}
</script>

<style scoped>

</style>