import service from '@/utils/request'

export const getContainerList = (data) => {
    return service({
        url: '/docker/containers',
        method: 'post',
    })
}

export const getImagesList = (data) => {
    return service({
        url: '/docker/images',
        method: 'post',
    })
}

export const getContainerStats = (containerId) => {
    return service({
        url: `/docker/containers/${containerId}/stats`,
        method: 'post',
    })
}

export const startContainer = (containerId) => {
    return service({
        url: `/docker/containers/${containerId}/start`,
        method: 'post',
    })
}

export const stopContainer = (containerId) => {
    return service({
        url: `/docker/containers/${containerId}/stop`,
        method: 'post',
    })
}

export const removeContainer = (containerId) => {
    return service({
        url: `/docker/containers/${containerId}/remove`,
        method: 'post',
    })
}

export const createAnacondaContainer = (jupyterPort, sshPort) => {
    return service({
        url: `/docker/containers/create_anaconda_container`,
        method: 'post',
        body: new URLSearchParams({
            jupyter_port: jupyterPort,
            ssh_port: sshPort,
        }),
    })
}