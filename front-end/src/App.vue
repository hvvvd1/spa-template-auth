<template>
  <Header />
  <div>
    <router-view />
  </div>
  <Footer />
</template>

<script>
import Header from "@/components/app1/Header";
import Footer from "@/components/app1/Footer";
import {store} from "@/components/store";

const getCookie = (name) => {
  return document.cookie.split("; ").reduce((r,v) => {
    const parts = v.split("=")
    return parts[0] === name ? decodeURIComponent(parts[1]) : r;
  }, "");
}

export default {
  components: {
    Header, Footer
  },
  data() {
    return {

    }
  },
  beforeMount() {
    let data = getCookie("_site_data")

    if (data !== "") {
      let cookieData = JSON.parse(data)

      // update store
      store.token = cookieData.token.token
      store.user = {
        id: cookieData.user.id,
        first_name: cookieData.user.first_name,
        last_name: cookieData.user.last_name,
        email: cookieData.user.email,
      }
    }
  }
}

</script>
