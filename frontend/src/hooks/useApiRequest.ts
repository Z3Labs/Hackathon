import { useState } from 'react'
import toast from 'react-hot-toast'

// 统一的API请求Hook
export const useApiRequest = () => {
  const [loading, setLoading] = useState(false)

  const request = async <T>(
    apiCall: () => Promise<T>,
    options?: {
      successMessage?: string
      errorMessage?: string
      showSuccessToast?: boolean
      showErrorToast?: boolean
    }
  ): Promise<T | null> => {
    const {
      successMessage = '操作成功',
      errorMessage = '操作失败',
      showSuccessToast = false,
      showErrorToast = true
    } = options || {}

    setLoading(true)
    
    try {
      const result = await apiCall()
      
      if (showSuccessToast) {
        toast.success(successMessage)
      }
      
      return result
    } catch (error) {
      console.error('API请求错误:', error)
      
      if (showErrorToast) {
        const message = error instanceof Error ? error.message : errorMessage
        toast.error(message)
      }
      
      return null
    } finally {
      setLoading(false)
    }
  }

  return {
    request,
    loading
  }
}

export default useApiRequest
