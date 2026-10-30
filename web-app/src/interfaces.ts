interface Channel {
    id: string
    name: string
    provider: string
    leagues: League[]
    language: string
    icon_url: string
}

interface League {
    id: string
    slug: string
    name: string
    region: string
    image: string
    priority: number
}