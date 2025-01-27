
export function Add(data=null) {
    return new Promise((resolve, reject) => {
        axios.post("/autoapi/Add", data).then(res=>{
            if (res.data.ok) {
                resolve(res.data.data)
            } else {
                reject(res.data.error)
            }
        }).then(error=>{
            reject(error)
        })
    })
}
