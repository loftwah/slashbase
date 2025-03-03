import type { NextPage } from 'next'
import { useRouter } from 'next/router'
import React, { useEffect, useState } from 'react'
import AppLayout from '../../../components/layouts/applayout'
import { Project } from '../../../data/models'
import DefaultErrorPage from 'next/error'
import { useAppDispatch, useAppSelector } from '../../../redux/hooks'
import { selectProjects } from '../../../redux/projectsSlice'
import { AddDBConnPayload } from '../../../network/payloads'
import { DBConnectionUseSSHType, ProjectMemberRole } from '../../../data/defaults'
import { addNewDBConn } from '../../../redux/allDBConnectionsSlice'
import Constants from '../../../constants'

const NewDBPage: NextPage = () => {

    const router = useRouter()
    const { id } = router.query

    const dispatch = useAppDispatch()
    const projects: Project[] = useAppSelector(selectProjects)
    const project = projects.find(x => x.id === id)

    const [dbName, setDBName] = useState('')
    const [dbHost, setDBHost] = useState('')
    const [dbPort, setDBPort] = useState('5432')
    const [dbDatabase, setDBDatabase] = useState('')
    const [dbUsername, setDBUsername] = useState('')
    const [dbPassword, setDBPassword] = useState('')
    const [dbUseSSH, setUseSSH] = useState<string>(DBConnectionUseSSHType.NONE)
    const [dbSSHHost, setSSHHost] = useState('')
    const [dbSSHUser, setSSHUser] = useState('')
    const [dbSSHPassword, setSSHPassword] = useState('')
    const [dbSSHKeyFile, setSSHKeyFile] = useState('')
    const [adding, setAdding] = useState(false)

    if (!project) {
        return <DefaultErrorPage statusCode={404} />
    }

    if (project.currentMember?.role !== ProjectMemberRole.ADMIN){
        return <DefaultErrorPage statusCode={401} title="Unauthorized" />
    }

    const startAddingDB = async () => {
      setAdding(true)
      const payload: AddDBConnPayload = {
          projectId: project.id,
          name: dbName,
          host: dbHost,
          port: dbPort,
          password: dbPassword,
          user: dbUsername,
          dbname: dbDatabase,
          useSSH: dbUseSSH,
          sshHost: dbSSHHost,
          sshUser: dbSSHUser,
          sshPassword: dbSSHPassword,
          sshKeyFile: dbSSHKeyFile,
          
      }
      await dispatch(addNewDBConn(payload))
      setAdding(false)
      router.replace(Constants.APP_PATHS.PROJECT.path.replace('[id]', project.id))
    }

    return (
        <AppLayout title="Add New Database Connection | Slashbase">
            <h1>Add New Postgres Database Connection in {project.name}</h1>
            <div className="form-container"> 
                <div className="field">
                    <label className="label">Display Name this database:</label>
                    <div className="control">
                        <input 
                            className="input" 
                            type="text" 
                            value={dbName}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setDBName(e.target.value)}}
                            placeholder="Enter a name for this database" />
                    </div>
                </div>
                <div className="field">
                    <label className="label">Host:</label>
                    <div className="control">
                        <input 
                            className="input" 
                            type="text" 
                            value={dbHost}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setDBHost(e.target.value)}}
                            placeholder="Enter host" />
                    </div>
                </div>
                <div className="field">
                    <label className="label">Port:</label>
                    <div className="control">
                        <input 
                            className="input" 
                            type="text" 
                            value={dbPort}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setDBPort(e.target.value)}}
                            placeholder="Enter port" />
                    </div>
                </div>
                <div className="field">
                    <label className="label">Database:</label>
                    <div className="control">
                        <input 
                            className="input" 
                            type="text" 
                            value={dbDatabase}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setDBDatabase(e.target.value)}}
                            placeholder="Enter database" />
                    </div>
                </div>
                <div className="field">
                    <label className="label">Database User:</label>
                    <div className="control">
                        <input 
                            className="input" 
                            type="text" 
                            value={dbUsername}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setDBUsername(e.target.value)}}
                            placeholder="Enter database username" />
                    </div>
                </div>
                <div className="field">
                    <label className="label">Database Password:</label>
                    <div className="control">
                        <input 
                            className="input" 
                            type="password" 
                            value={dbPassword}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setDBPassword(e.target.value)}}
                            placeholder="Enter database password" />
                    </div>
                </div>
                <div className="field">
                    <label className="label">Use SSH:</label>
                    <div className="select">
                        <select
                            value={dbUseSSH}
                            onChange={(e: React.ChangeEvent<HTMLSelectElement>)=>{
                                setUseSSH(e.target.value)
                            }}
                        >
                            <option 
                                value={DBConnectionUseSSHType.NONE}>
                                None
                            </option>
                            <option 
                                value={DBConnectionUseSSHType.PASSWORD}>
                                Password
                            </option>
                            <option 
                                value={DBConnectionUseSSHType.KEYFILE}>
                                Identity File
                            </option>
                            <option 
                                value={DBConnectionUseSSHType.PASSKEYFILE}>
                                Identity File with Password
                            </option>
                        </select>
                    </div>
                </div>
                { dbUseSSH !== DBConnectionUseSSHType.NONE && 
                    <React.Fragment>
                        <div className="field">
                            <label className="label">SSH Host:</label>
                            <div className="control">
                                <input 
                                    className="input" 
                                    type="text" 
                                    value={dbSSHHost}
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setSSHHost(e.target.value)}}
                                    placeholder="Enter SSH Host" />
                            </div>
                        </div>
                        <div className="field">
                            <label className="label">SSH User:</label>
                            <div className="control">
                                <input 
                                    className="input" 
                                    type="text" 
                                    value={dbSSHUser}
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setSSHUser(e.target.value)}}
                                    placeholder="Enter SSH User" />
                            </div>
                        </div>
                        { (dbUseSSH === DBConnectionUseSSHType.PASSWORD || dbUseSSH === DBConnectionUseSSHType.PASSKEYFILE ) &&
                            < div className="field">
                                <label className="label">SSH Password:</label>
                                <div className="control">
                                    <input 
                                        className="input" 
                                        type="password" 
                                        value={dbSSHPassword}
                                        onChange={(e: React.ChangeEvent<HTMLInputElement>)=>{setSSHPassword(e.target.value)}}
                                        placeholder="Enter SSH Password" />
                                </div>
                            </div> 
                        }
                        { (dbUseSSH === DBConnectionUseSSHType.KEYFILE || dbUseSSH === DBConnectionUseSSHType.PASSKEYFILE ) &&
                            <div className="field">
                                <label className="label">SSH Identity File:</label>
                                <div className="control">
                                    <textarea 
                                        className="textarea" 
                                        value={dbSSHKeyFile}
                                        onChange={(e: React.ChangeEvent<HTMLTextAreaElement>)=>{setSSHKeyFile(e.target.value)}}
                                        placeholder="Paste the contents of SSH Identity File here"/>
                                </div>
                            </div>
                        }
                    </React.Fragment>
                }
                <div className="control">
                    { !adding && <button className="button is-primary" onClick={startAddingDB}>Add</button> }
                    { adding && <button className="button is-primary">Adding...</button>}
                </div>
            </div>
        </AppLayout>
    )
}

export default NewDBPage
