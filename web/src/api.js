import axios from 'axios';

class Api {
  constructor(props) {
    this.client = axios.create({
      baseURL: props
    })
  }
  async policies() {
    try {
      const res = await this.client.get('/policies')
      return res.data
    } catch (e) {
      if (e.response && e.response.data) {
        throw new Error(e.response.message);
      }
    }
    return 
  }
  async addPolicy(policy, before) {
    console.log(policy)
    try {
      const res = await this.client.post(`/policies`, policy, {
        params: { before }
      })
      return res.data
    } catch (e) {
      if (e.response && e.response.data) {
        throw new Error(e.response.message);
      }
    }
    return 
  }
  async modifyPolicy(id, policy, before) {
    try {
      const res = await this.client.put(`/policies/${id}`, policy, {
        params: { before }
      })
      return res.data
    } catch (e) {
      if (e.response && e.response.data) {
        throw new Error(e.response.message);
      }
    }
    return 
  }
  async removePolicy(id) {
    try {
      const res = await this.client.delete(`/policies/${id}`)
      return res.data
    } catch (e) {
      if (e.response && e.response.data) {
        throw new Error(e.response.message);
      }
    }
    return 
  }
};

export default new Api('/api/v1');
