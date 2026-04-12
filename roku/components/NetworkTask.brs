sub init()
    m.top.functionName = "executeRequest"
end sub

sub executeRequest()
    url = m.top.url
    if url = "" then return

    utrans = CreateObject("roUrlTransfer")
    utrans.SetCertificatesFile("common:/certs/ca-bundle.crt")
    utrans.InitClientCertificates()
    utrans.SetUrl(url)

    responseStr = utrans.GetToString()
    if responseStr <> ""
        json = ParseJson(responseStr)
        if json <> invalid and json.items <> invalid
            root = CreateObject("roSGNode", "ContentNode")
            for each item in json.items
                node = root.CreateChild("ContentNode")
                node.title = item.title
                node.description = item.description
                node.hdPosterUrl = item.hdPosterUrl
                node.id = item.id
            end for
            m.top.content = root
        else
            m.top.error = "Failed to parse Response"
        end if
    else
        m.top.error = "Network Request Failed"
    end if
end sub
