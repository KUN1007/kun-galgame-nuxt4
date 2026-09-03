export const useTabQuery = (defaultValue: string, key = 'tab') => {
  const route = useRoute()
  const router = useRouter()

  return computed<string>({
    get() {
      const v = route.query[key]
      const val = Array.isArray(v) ? v[0] : v
      return val || defaultValue
    },
    set(val) {
      const { [key]: _dropped, ...rest } = route.query
      router.replace({
        query: val === defaultValue ? rest : { ...rest, [key]: val }
      })
    }
  })
}
