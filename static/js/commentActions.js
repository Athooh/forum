// likes and  dislike buttons functionalities
document.addEventListener('DOMContentLoaded', () => {
    // Only add event listeners if user is authenticated
    const likeButtons = document.querySelectorAll('.btn-like');
    const dislikeButtons = document.querySelectorAll('.btn-dislike');

    if (likeButtons.length > 0) {
        // Function to handle reaction response
        async function handleReactionResponse(response, postId) {
            const contentType = response.headers.get('content-type');
            if (!contentType || !contentType.includes('application/json')) {
                // If response is not JSON, show guest prompt for specific comment
                const promptContainer = document.getElementById(`reaction-prompt-${commentId}`);
                if (promptContainer) {
                    promptContainer.style.display = 'block';
                }
                return null;
            }
            return await response.json();
        }

        // Handle Like Button Click
        likeButtons.forEach(button => {
            button.addEventListener('click', async () => {
                const postId = button.getAttribute('data-id');
                const promptContainer = document.getElementById(`reaction-prompt-${commentId}`);
                
                // Check if user is authenticated
                const isAuthenticated = document.querySelector('.user-info') !== null;
                
                if (!isAuthenticated) {
                    // Show guest prompt if not authenticated
                    promptContainer.style.display = 'block';
                    return;
                }

                try {
                    const response = await fetch(`/like-comment?id=${commentId}`, { method: 'POST' });
                    const data = await handleReactionResponse(response, commentId);
                    if (data) {
                        updateReactionCounts(commentId, data.likes, data.dislikes);
                    }
                } catch (err) {
                    console.error('Error:', err);
                }
            });
        });

        // Handle Dislike Button Click
        dislikeButtons.forEach(button => {
            button.addEventListener('click', async () => {
                const postId = button.getAttribute('data-id');
                const promptContainer = document.getElementById(`reaction-prompt-${commentId}`);
                
                // Check if user is authenticated
                const isAuthenticated = document.querySelector('.user-info') !== null;
                
                if (!isAuthenticated) {
                    // Show guest prompt if not authenticated
                    promptContainer.style.display = 'block';
                    return;
                }

                try {
                    const response = await fetch(`/dislike-comment?id=${commentId}`, { method: 'POST' });
                    const data = await handleReactionResponse(response, commentId);
                    if (data) {
                        updateReactionCounts(commentId, data.likes, data.dislikes);
                    }
                } catch (err) {
                    console.error('Error:', err);
                }
            });
        });

        // Close prompts when clicking outside
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.btn-like') && 
                !e.target.closest('.btn-dislike') && 
                !e.target.closest('.guest-reaction-prompt')) {
                document.querySelectorAll('.reaction-prompt-container').forEach(container => {
                    container.style.display = 'none';
                });
            }
        });
    }

    // Handle post options dropdown
    const optionsButtons = document.querySelectorAll('.btn-options');
    
    optionsButtons.forEach(button => {
        button.addEventListener('click', (e) => {
            e.stopPropagation();
            const dropdown = button.nextElementSibling;
            
            // Close all other dropdowns first
            document.querySelectorAll('.options-dropdown').forEach(d => {
                if (d !== dropdown) {
                    d.style.display = 'none';
                }
            });
            
            // Toggle current dropdown
            dropdown.style.display = dropdown.style.display === 'none' ? 'block' : 'none';
        });
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.post-options-menu')) {
            document.querySelectorAll('.options-dropdown').forEach(dropdown => {
                dropdown.style.display = 'none';
            });
        }
    });

    // Add confirmation for delete action
    document.querySelectorAll('.options-dropdown form').forEach(form => {
        form.addEventListener('submit', (e) => {
            if (!confirm('Are you sure you want to delete this post?')) {
                e.preventDefault();
            }
        });
    });
});

// Function to update all instances of a comments's reaction counts
function updateReactionCounts(commentId, likes, dislikes) {
    // Update in dashboard view
    const dashboardLikeButtons = document.querySelectorAll(`.btn-like[data-id="${commentId}"]`);
    const dashboardDislikeButtons = document.querySelectorAll(`.btn-dislike[data-id="${commentId}"]`);

    dashboardLikeButtons.forEach(button => {
        button.querySelector('span').innerText = likes;
    });

    dashboardDislikeButtons.forEach(button => {
        button.querySelector('span').innerText = dislikes;
    });

    // Update in view-post view if we're on that page
    const viewCommentLikeButton = document.querySelector(`.post-actions .btn-like[data-id="${commentId}"] span`);
    const viewCommentDislikeButton = document.querySelector(`.post-actions .btn-dislike[data-id="${commentId}"] span`);

    if (viewCommentLikeButton) viewCommentLikeButton.innerText = likes;
    if (viewCommentDislikeButton) viewCommentDislikeButton.innerText = dislikes;
}
