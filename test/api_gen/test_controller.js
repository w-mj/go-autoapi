
export function TestController_int_none(data=null) {
    return new Promise((resolve, reject) => {
        axios.post("/autoapi/TestController_int_none", data).then(res=>{
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

export function TestController_UserInfo_UserInfo(data=null) {
    return new Promise((resolve, reject) => {
        axios.post("/autoapi/TestController_UserInfo_UserInfo", data).then(res=>{
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

export function TestController_UserInfo_UserInfo_2(data=null) {
    return new Promise((resolve, reject) => {
        axios.post("/autoapi/TestController_UserInfo_UserInfo_2", data).then(res=>{
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

export function TestController_UserInfo_UserInfo_3(data=null) {
    return new Promise((resolve, reject) => {
        axios.post("/autoapi/TestController_UserInfo_UserInfo_3", data).then(res=>{
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
