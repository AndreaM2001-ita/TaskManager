import React, { useState, useEffect } from "react";
import Input from "./components/Input";
import Button from "./components/Button";
import Card from "./components/Card";
import "./TaskManager.css";

function TaskManager() {
  const [tasks, setTasks] = useState([]);
  const [search, setSearch] = useState("");
  const [task, setTask] = useState({ name: "", description: "" });
  const [update, setUpdate] =useState(null);

  const [errorMsg, setErrorMsg] = useState("");
  //get all the tasks
  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const response = await fetch("http://localhost:8080/api/tasks");
        if (!response.ok) throw new Error("cannot fetch tasks");
        const tasks = await response.json();
        setTasks(tasks); // Set fetched tasks to the state
      } catch (error) {
        setErrorMsg("Unable to fetch tasks. Please try again later.");
      }
    };

    fetchTasks(); // Call the fetch function when the component mounts
  }, []); 

  //use async as program does not need to halt for response
  const addTask = async () => {
    if (!task.name.trim()) return;  // if task empty, do not add it
    //create a new task
    const newTask={
      ...task,
      customId: Date.now().toString(), //unique id
      dateCreated: new Date().toLocaleString(), //time equal to id but human readable
    };

    try{
      //POSt to backend - send new task
      const response = await fetch("http://localhost:8080/api/tasks",{
        method: "POST",
        headers: {
          "Content-Type": "application/json", //sending json data
        },
        body: JSON.stringify(newTask),  //send content oif new task
      })
      if(!response.ok) throw new Error("Error... Failed to save task");  //problems with server go

      const saveTask = await response.json();
      console.log(saveTask)
      setTasks(prevTasks => {
        if (!Array.isArray(prevTasks)) {
          console.error("prevTasks is not an array, resetting to empty array.");
          prevTasks = [];  // Fallback to an empty array if it's not iterable
        }
        return [...prevTasks, saveTask];
       });// Add new task to the list);//save in local array the new task
      setTask({ name: "", description: "" });  //reset to blan the input fields
      setErrorMsg("");// reset error message when needed
    }
    catch(error){
      console.error("Error saving task", error);
      setErrorMsg("Server not available. Please try again later.");
    }
  };

  //async as rogram should not wait for the response
  const deleteTask = async (customId) => {
    setTasks(tasks.filter((task) => task.customId !== customId));
    console.log(customId)
    try {
      //send DELETE
      const response = await fetch(`http://localhost:8080/api/tasks/${customId}`, {
        method: "DELETE",
      });
  
      if (!response.ok) throw new Error("Error... Failed to delete task");
    } catch (error) {
      console.error("Error deleting task", error);
      setErrorMsg("Error deleting task");
      
      //Rollback in case of error
      setTasks([...tasks]);
    }
  };

  const updateTask = async (customId) => {
      const updatedData = {
        ...task,
        customId: update,
        dateCreated: new Date().toLocaleString(), 
      };
      
      try {
        const response = await fetch(`http://localhost:8080/api/tasks/${update}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(updatedData),
        });
    
        if (!response.ok) throw new Error("Error... Failed to update task");

        const updatedTask = await response.json();
        setTasks(tasks.map(task =>
          task.customId === update ? {...updatedTask } : task
        ));
        setErrorMsg(""); // clear any previous error
        setTask({ name: "", description: "" })
        setUpdate(null)

      } catch (error) {
        console.error("Error updating task", error);
        setErrorMsg("Error updating task");
      }
  };
  const saveUpdatedTask = (customId) => {
    const taskToEdit = tasks.find(t => t.customId === customId);
    if (taskToEdit) {
      setTask({ name: taskToEdit.name, description: taskToEdit.description });
      setUpdate(customId);
    }
  };

  return (
    <div className="task-manager-container">
      <h1 className="task-manager-title">Task Manager</h1>

      <div className="search-container">
        <Input
          placeholder="Search tasks..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="input-field"
        />
      </div>

      <div className="task-inputs">
        <Input
          placeholder="Task Name"
          value={task.name}
          onChange={(e) => setTask({ ...task, name: e.target.value })}
          className="input-field"
        />
        <Input
          placeholder="Task Description"
          value={task.description}
          onChange={(e) => setTask({ ...task, description: e.target.value })}
          className="input-field"
        />
        <Button
          onClick={update ? updateTask : addTask}
          className="add-task-button"
        >
          {update ? "Update Task" : "Add Task"}
        </Button>
        {errorMsg && <div className="error-message">{errorMsg}</div>}
      </div>

      <div className="task-list">
        {(tasks || []).filter((t) => t.name.toLowerCase().includes(search.toLowerCase()))
          .map((t) => (
            <Card key={t.customId} className="task-card">
              <div className="task-info">
                <h2 className="task-name">{t.name}</h2>
                <p className="task-description">{t.description}</p>
                <p className="task-date">Created: {t.dateCreated}</p>
              </div>
              <div className="task-actions">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => saveUpdatedTask(t.customId)}
                  className="edit-button"
                >
                  Edit
                </Button>
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => deleteTask(t.customId)}
                  className="delete-button"
                >
                  Delete
                </Button>
              </div>
            </Card>
          ))}
      </div>
    </div>
  );
}

export default TaskManager;
