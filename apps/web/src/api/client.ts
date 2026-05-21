import ky from 'ky'

const api = ky.create({
  prefixUrl: '/',
  credentials: 'include',
  hooks: {
    beforeRequest: [
      (request) => {
        const token = localStorage.getItem('gatewarden_token')
        if (token) {
          request.headers.set('Authorization', `Bearer ${token}`)
        }
      },
    ],
    afterResponse: [
      async (_request, _options, response) => {
        if (response.status === 401) {
          localStorage.removeItem('gatewarden_token')
          localStorage.removeItem('gatewarden_user')
          window.location.href = '/login'
        }
      },
    ],
  },
})

export { api }
