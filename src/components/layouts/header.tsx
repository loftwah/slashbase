import styles from './header.module.scss'
import React, { useEffect } from 'react'
import Link from 'next/link'
import { DBConnection, Project, User } from '../../data/models'
import { useAppSelector } from '../../redux/hooks'
import { useRouter } from 'next/router'
import { selectCurrentUser } from '../../redux/currentUserSlice'
import Constants from '../../constants'
import { selectProjects } from '../../redux/projectsSlice'
import { selectDBConnection } from '../../redux/dbConnectionSlice'

type HeaderPropType = {

}

const Header = (_: HeaderPropType) => {
    
    const router = useRouter()

    const currentUser: User = useAppSelector(selectCurrentUser)
    const projects: Project[] = useAppSelector(selectProjects)

    const options = [
        { value: 'home', label: 'Home', path: Constants.APP_PATHS.HOME.as },
        ...projects.map((x: Project) => ({value: x.id, label: x.name, path: Constants.APP_PATHS.PROJECT.as+x.id }))
    ]

    const onNavigate = (event: React.ChangeEvent<HTMLSelectElement>) => {
        router.replace(options.find(x=>x.value === event.target.value)!.path)
    }

    let currentOption = 'home'
    if (router.pathname === Constants.APP_PATHS.PROJECT.href) {
        currentOption = String(router.query.id)
    } else if (router.pathname === Constants.APP_PATHS.DB.href) {
        const currentDBConnection: DBConnection | undefined = useAppSelector(selectDBConnection)
        if (currentDBConnection)
            currentOption = currentDBConnection?.projectId
    }

    return (
        <header className={styles.header}>
            <Link {...Constants.APP_PATHS.HOME}>
                <a>
                    <div className={styles.home}>
                        <i className={"fas fa-home"}/>
                    </div>
                </a>
            </Link>
            <div className={styles.headerCenter}>
                <select className={styles.headerSelect} value={currentOption} onChange={onNavigate}>
                    {options.map((x)=>{
                        return <option key={x.value} value={x.value} label={x.label} />
                    })}
                </select>
            </div>
            <div className={styles.headerMenu}>
                { currentUser && 
                <Link {...Constants.APP_PATHS.LOGOUT}>
                    <a>
                        <img className={styles.profileImage} src={currentUser.profileImageUrl} width={40} height={40} /> 
                    </a>
                </Link>
                }
            </div>
        </header>
    )
}


export default Header