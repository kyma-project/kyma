<template>
  <footer class="my-footer">
    <p>© 2025 Kyma Project. All rights reserved.</p>
    <p>
      <a href="#">Privacy Policy</a> | <a href="#">Terms of Service</a>
    </p>
  </footer>
</template>

<script>
export default {
  name: 'MyFooter'
}
</script>

<style scoped>
.my-footer {
  text-align: center;
  padding: 2rem 0;
  border-top: 1px solid var(--vp-c-divider);
  margin-top: 2rem;
  color: var(--vp-c-text-2);
}
</style>
